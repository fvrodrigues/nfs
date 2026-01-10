package workflow

import (
	"errors"
	"fmt"
	"nfse/pkg/config"
	"nfse/pkg/logger"
	"nfse/pkg/receita"
	"nfse/pkg/rod"
	"nfse/pkg/sheets"
	"nfse/pkg/ui"
	"os"
)

type Workflow struct {
	logger   *logger.ArquivoLog
	cfg      config.Config
	planilha *sheets.Planilha
	pagina   *rod.Pagina
	receita  *receita.Receita
	ui       *ui.UI
}

func New(logger *logger.ArquivoLog, cfg config.Config, planilha *sheets.Planilha, pagina *rod.Pagina, receita *receita.Receita, ui *ui.UI) *Workflow {
	return &Workflow{
		logger:   logger,
		planilha: planilha,
		pagina:   pagina,
		receita:  receita,
		cfg:      cfg,
		ui:       ui,
	}
}

func (w *Workflow) Executar() error {
	defer w.logger.EncerrarAplicacao()

	w.ui.Workflow("Pegando dados das notas fiscais na planilha...")
	dadosNotasFiscais, err := w.planilha.PegarDadosDeNotasFiscais()
	if err != nil {
		return err
	}

	w.ui.Workflow("Coletando dados de login e vendo as notas que precisam ser emitidas...")
	clientes, err := w.planilha.ColetarDadosLogin(dadosNotasFiscais)
	if err != nil {
		return err
	}
	if len(clientes) == 0 {
		w.ui.Sucesso("Não há notas fiscais pendentes para emitir. Encerrando aplicação")
		return nil
	}

	fmt.Printf("\n\n")

	for _, cliente := range clientes {
		w.ui.WorkflowComArg("Começando processo de emissão para %v", cliente.Empresa)

		err := w.receita.AcessarSiteReceita(w.cfg.Website)
		if err != nil {
			if !errors.Is(err, receita.ErrSessaoAbortada) {
				return fmt.Errorf("erro ao acessar site receita: %w", err)
			}

			err = w.Retry(err,
				"acessar website da receita",
				func() error {
					return w.receita.AcessarSiteReceita(w.cfg.Website)
				})
			if err != nil {
				return err
			}
		}
		w.ui.MsgComArg("Acessado: %s", w.cfg.Website)

		err = w.receita.ApertarLoginUnico()
		if err != nil {
			if !errors.Is(err, receita.ErrNaoEncontrouElemento) {
				return fmt.Errorf("erro genérico ao apertar botão de login único: %w", err)
			}

			err = w.Retry(err,
				"encontrar botão login único",
				w.receita.ApertarLoginUnico)
			if err != nil {
				return err
			}
		}
		w.ui.Msg("Botão de login único encontrado. Indo para a página de login.")

		err = w.receita.FazerLogin(cliente.Login, cliente.Senha)
		if err != nil {
			switch {
			case errors.Is(err, receita.ErrDadosLoginInvalidos):
				w.ui.Erro(err)
				continue
			case errors.Is(err, receita.ErrNaoEncontrouElemento):
				err = w.Retry(err,
					"encontrar campo de login",
					func() error {
						return w.receita.FazerLogin(cliente.Login, cliente.Senha)
					})
				if err != nil {
					return err
				}
			default:
				return err
			}
		}
		w.ui.MsgComArg("Login feito para %v", cliente.Empresa)

		err = w.receita.ApertarBotaoEmissao()
		if err != nil {
			if !errors.Is(err, receita.ErrNaoCarregaNovaPagina) {
				return err
			}

			err = w.Retry(err,
				"encontrar botão para forms de emissão de NFSe",
				w.receita.ApertarBotaoEmissao)
			if err != nil {
				return err
			}

		}
		w.ui.Msg("Botão lateral para página de emissão de NFSE encontrado.")

		for i, nota := range cliente.NotasEmitir {
			w.ui.Msg(fmt.Sprintf("Emitindo nota %d de %d", i+1, len(cliente.NotasEmitir)))
			if err := w.receita.ColocaCnpjEData(nota.Cnpj, nota.Data); err != nil {
				if !errors.Is(err, receita.ErrNaoEncontrouElemento) {
					return fmt.Errorf("erro genérico ao colocar data/cnpj: %w", err)
				}

				err = w.Retry(err,
					fmt.Sprintf("colocar data/cnpj na nota %d", i+1),
					func() error {
						return w.receita.ColocaCnpjEData(nota.Cnpj, nota.Data)
					})
				if err != nil {
					return err
				}
			}

			if true {
				w.ui.Msg("debug acabou")
				os.Exit(0)
			}
		}

		w.ui.SucessoComArg("Todas as notas fiscais para %s foram emitidas com sucesso!", cliente.Empresa)

		w.ui.Msg("Deslogando do site...")
		err = w.receita.Deslogar()
		if err != nil {
			if !errors.Is(err, receita.ErrNaoEncontrouElemento) && !errors.Is(err, receita.ErrNaoCarregaNovaPagina) {
				return fmt.Errorf("erro genérico ao deslogar: %w", err)
			}

			err = w.Retry(err,
				"deslogar do site",
				w.receita.Deslogar)
			if err != nil {
				w.ui.Msg("Voltando ao início sem deslogar, o que pode fazer o site desconfiar da automação")
				continue
			}
		}
	}
	w.ui.Sucesso("Todas as notas fiscais foram emitidas com sucesso!")
	return nil
}

func (w *Workflow) Retry(erro error, operacao string, fn func() error) error {
	w.ui.Erro(erro)
	w.ui.Msg("Tentando novamente...")

	for i := 1; i <= 3; i++ {
		if err := fn(); err != nil {
			w.ui.MsgComArg("Erro na Tentativa %d", i)
			if i == 3 {
				return fmt.Errorf("falha definitiva em %s: %v", operacao, err)
			}
			continue
		}
		w.ui.Sucesso(operacao)
		return nil
	}
	return nil
}
