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
	if clientes == nil {
		w.ui.Sucesso("Não há notas fiscais pendentes para emitir. Encerrando aplicação")
		return nil
	}

	fmt.Printf("\n\n")

	for _, cliente := range clientes {
		w.ui.WorkflowComArg("Começando processo de emissão para %v", cliente.Empresa)

		if err := w.receita.AcessarSiteReceita(w.cfg.Website); err != nil {
			return err
		}
		w.ui.MsgComArg("Acessado: %s", w.cfg.Website)

		if err := w.receita.ApertarLoginUnico(); err != nil {
			switch {
			case errors.Is(err, receita.ErrNaoEncontrouElemento):
				err := w.Retry(err, "encontrar botão login único", func() error {
					return w.receita.ApertarLoginUnico()
				})
				if err != nil {
					return err
				}
			default:
				return err
			}
		}
		w.ui.Msg("Botão de login encontrado")

		if err := w.receita.FazerLogin(cliente.Login, cliente.Senha); err != nil {
			switch {
			case errors.Is(err, receita.ErrDadosLoginInvalidos):
				w.ui.Erro(err)
				continue
			case errors.Is(err, receita.ErrNaoEncontrouElemento):
				if err := w.Retry(err,
					"encontrar campo de login",
					func() error {
						return w.receita.FazerLogin(cliente.Login, cliente.Senha)
					}); err != nil {
					return err
				}
			default:
				return err
			}
		}

		w.ui.MsgComArg("Login feito para %v", cliente.Empresa)

		if err := w.receita.ApertarBotaoEmissao(); err != nil {
			if !errors.Is(err, receita.ErrNaoCarregaNovaPagina) {
				return err
			}

			if err := w.Retry(err,
				"Encontrar botão para forms de emissão de NFSe",
				w.receita.ApertarBotaoEmissao); err != nil {
				return err
			}

		}
		w.ui.Msg("Botão para página de emissão encontrado.")

		if err := w.receita.Deslogar(); err != nil {
			return err
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
