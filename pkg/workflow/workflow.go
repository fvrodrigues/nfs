package workflow

import (
	"errors"
	"fmt"
	"nfse/pkg/config"
	"nfse/pkg/domain"
	"nfse/pkg/logger"
	"nfse/pkg/receita"
	"nfse/pkg/rod"
	"nfse/pkg/sheets"
	"nfse/pkg/sistema"
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

func New(logger *logger.ArquivoLog, cfg config.Config, planilha *sheets.Planilha, ui *ui.UI) *Workflow {
	return &Workflow{
		logger:   logger,
		planilha: planilha,
		cfg:      cfg,
		ui:       ui,
	}
}

func (w *Workflow) Executar(prestador domain.Prestador) error {
	if err := prestador.ValidaDadosCorpoReq(); err != nil {
		return err
	}

	pathPrestador, err := sistema.CriarPastaParaPrestador(prestador.Prestador)
	if err != nil {
		return fmt.Errorf("erro ao criar pasta para prestador: %w", err)
	}
	fmt.Printf("Criada pasta para prestador %s: %s\n", prestador.Prestador, pathPrestador)

	pagina, err := rod.CriarNavegador(w.logger, false)
	if err != nil {
		return fmt.Errorf("erro ao criar navegador: %w", err)
	}
	defer pagina.Close()

	pagReceita := receita.New(pagina)
	err = pagReceita.DefinirPastaDownload(pathPrestador)
	if err != nil {
		err = w.Retry(err,
			"configurar pasta de Download para prestador",
			func() error { return w.pagina.DefinirPastaDownload(pathPrestador) })
		if err != nil {
			return err
		}
	}
	w.ui.WorkflowComArg("Começando processo de emissão para %v", prestador.Prestador)

	err = pagReceita.AcessarSiteReceita(w.cfg.Website)
	if err != nil {
		if !errors.Is(err, receita.ErrSessaoAbortada) {
			return fmt.Errorf("erro ao acessar site receita: %w", err)
		}

		err = w.Retry(err,
			"acessar website da receita",
			func() error {
				return pagReceita.AcessarSiteReceita(w.cfg.Website)
			})
		if err != nil {
			return err
		}
	}
	w.ui.MsgComArg("Acessado: %s", w.cfg.Website)

	err = pagReceita.ApertarLoginUnico()
	if err != nil {
		if !errors.Is(err, receita.ErrNaoEncontrouElemento) {
			return fmt.Errorf("erro genérico ao apertar botão de login único: %w", err)
		}

		err = w.Retry(err,
			"encontrar botão login único",
			pagReceita.ApertarLoginUnico)
		if err != nil {
			return err
		}
	}
	w.ui.Msg("Botão de login único encontrado. Indo para a página de login.")

	err = pagReceita.FazerLogin(prestador.Login, prestador.Senha)
	if err != nil {
		switch {
		case errors.Is(err, receita.ErrDadosLoginInvalidos):
			w.ui.Erro(err)
			return fmt.Errorf("%w: dados de login inválidos para %s", err, prestador.Prestador)
		case errors.Is(err, receita.ErrNaoEncontrouElemento):
			err = w.Retry(err,
				"encontrar campo de login",
				func() error {
					return pagReceita.FazerLogin(prestador.Login, prestador.Senha)
				})
			if err != nil {
				return err
			}
		default:
			return err
		}
	}
	w.ui.MsgComArg("Login feito para %v", prestador.Prestador)

	err = pagReceita.IrParaFormsEmissao()
	if err != nil {
		if !errors.Is(err, receita.ErrNaoCarregaNovaPagina) {
			return err
		}

		err = w.Retry(err,
			"encontrar botão para forms de emissão de NFSe",
			pagReceita.IrParaFormsEmissao)
		if err != nil {
			return err
		}

	}
	w.ui.Msg("Botão lateral para página de emissão de NFSE encontrado.")

	for i, nota := range prestador.NotasEmitir {
		w.ui.Msg(fmt.Sprintf("Emitindo nota %d de %d", i+1, len(prestador.NotasEmitir)))
		if err := pagReceita.ColocaCnpjEData(nota.Cnpj, nota.Data); err != nil {
			if !errors.Is(err, receita.ErrNaoEncontrouElemento) {
				return fmt.Errorf("erro genérico ao colocar data/cnpj: %w", err)
			}

			err = w.Retry(err,
				fmt.Sprintf("colocar data/cnpj na nota %d", i+1),
				func() error {
					return pagReceita.ColocaCnpjEData(nota.Cnpj, nota.Data)
				})
			if err != nil {
				return err
			}
		}

		if err := pagReceita.ColocarDadosEEmitirNF(nota.Tomador, nota.Observacao, nota.Valor); err != nil {
			if !errors.Is(err, receita.ErrNaoEncontrouElemento) || !errors.Is(err, receita.ErrNaoCarregaNovaPagina) {
				return fmt.Errorf("erro genérico ao colocar data/cnpj: %w", err)
			}

			err = w.Retry(err,
				fmt.Sprintf("colocar dados da nota %d", i+1),
				func() error {
					return pagReceita.ColocarDadosEEmitirNF(nota.Tomador, nota.Observacao, nota.Valor)
				})
		}

		if err := pagReceita.VoltarParaFormDeNota(); err != nil {
			if !errors.Is(err, receita.ErrNaoCarregaNovaPagina) || !errors.Is(err, receita.ErrNaoEncontrouElemento) {
				return fmt.Errorf("erro genérico ao voltar para form de nota: %w", err)
			}

			err = w.Retry(err,
				"voltar para form de nota %d",
				pagReceita.VoltarParaFormDeNota)
			if err != nil {
				return err
			}
		}
	}

	w.ui.SucessoComArg("Todas as notas fiscais para %s foram emitidas com sucesso!", prestador.Prestador)

	//w.ui.Msg("Deslogando do site...")
	//err = pagReceita.Deslogar()
	//if err != nil {
	//	if !errors.Is(err, receita.ErrNaoEncontrouElemento) && !errors.Is(err, receita.ErrNaoCarregaNovaPagina) {
	//		return fmt.Errorf("erro genérico ao deslogar: %w", err)
	//	}
	//
	//	err = w.Retry(err,
	//		"deslogar do site",
	//		pagReceita.Deslogar)
	//	if err != nil {
	//		w.ui.Msg("Voltando ao início sem deslogar, o que pode fazer o site desconfiar da automação")
	//		continue
	//	}
	//}

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
