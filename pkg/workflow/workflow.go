package workflow

import (
	"errors"
	"fmt"
	"nfse/pkg/captcha"
	"nfse/pkg/domain"
	"nfse/pkg/logger"
	"nfse/pkg/receita"
	"nfse/pkg/rod"
	"nfse/pkg/sistema"
	"nfse/pkg/ui"
)

type Workflow struct {
	logger  *logger.ArquivoLog
	captcha *captcha.Captcha
	pagina  *rod.Pagina
	receita *receita.Receita
	ui      *ui.UI
}

func New(logger *logger.ArquivoLog, captcha *captcha.Captcha, ui *ui.UI) *Workflow {
	return &Workflow{
		logger:  logger,
		captcha: captcha,
		ui:      ui,
	}
}

func (w *Workflow) Executar(prestador domain.Prestador, reqID string) error {
	if err := prestador.ValidaDadosCorpoReq(); err != nil {
		w.escreverErro(true, reqID, err)
		return err
	}
	w.escreverMsg(true, reqID, "Dados de requisição recebidos, começando emissão para %s", prestador.Prestador)

	pathPrestador, err := sistema.CriarPastaParaPrestador(prestador.Prestador)
	if err != nil {
		w.escreverErro(true, reqID, "Erro ao criar pasta para prestador: %v", err)
		return fmt.Errorf("erro ao criar pasta para prestador: %w", err)
	}
	w.escreverMsg(false, "", "Pasta para prestador %s pronta: %s\n", prestador.Prestador, pathPrestador)

	pagina, err := rod.CriarNavegador(false)
	if err != nil {
		w.escreverErro(true, reqID, "v: %v", rod.ErrCriarNavegador, err)
		return fmt.Errorf("%w: %w", rod.ErrCriarNavegador, err)
	}
	w.escreverMsg(true, reqID, "Instância de navegador criada")

	pagReceita := receita.New(pagina, w.captcha)
	err = pagReceita.DefinirPastaDownload(pathPrestador)
	if err != nil {
		w.escreverErro(true, reqID, "v: %v", rod.ErrConfigurarPastaDownloadDefault, err)

		err = w.retry(err, reqID,
			"configurar pasta de Download para prestador",
			func() error { return w.pagina.DefinirPastaDownload(pathPrestador) })
		if err != nil {
			return err
		}
	}
	w.escreverMsg(true, reqID, "Começando processo de emissão para %s\n", prestador.Prestador)

	err = pagReceita.AcessarSiteReceita()
	if err != nil {
		w.escreverErro(true, reqID, "erro ao acessar site da receita: %v", err)

		if errors.Is(err, receita.ErrSessaoAbortada) {
			w.escreverErro(true, reqID, "%v: motivo desconhecido: %v", receita.ErrSessaoAbortada, err)
			return fmt.Errorf("sessão abortada ao acessar site receita: %w", err)
		}

		err = w.retry(err, reqID,
			"acessar site da receita",
			func() error {
				return pagReceita.AcessarSiteReceita()
			})
		if err != nil {
			return err
		}
	}
	w.escreverMsg(true, reqID, "Acessado site da receita.")

	err = pagReceita.EncontrarBotaoLoginUnico()
	if err != nil {
		w.escreverErro(true, reqID, "erro ao achar botão de login único: %v", err)
		err = w.retry(err, reqID,
			"encontrar botão de login único",
			func() error { return pagReceita.EncontrarBotaoLoginUnico() })
		if err != nil {
			return err
		}
	}
	err = pagReceita.ApertarLoginUnico()
	if err != nil {
		w.escreverErro(true, reqID, "erro ao apertar botão de login único apesar de ter sido encontrado na página: %v", err)

		err = w.retry(err, reqID,
			"encontrar botão login único",
			pagReceita.ApertarLoginUnico)
		if err != nil {
			return err
		}
	}
	w.escreverMsg(false, "", "Botão de login único encontrado.")

	fmt.Printf(pagina.HTML())

	err = pagReceita.ColocarDadosLogin(prestador.Login, prestador.Senha)
	if err != nil {
		w.escreverErro(true, reqID, "erro ao colocar dados de login: %v", err)
		if errors.Is(err, receita.ErrNaoEncontrouElemento) {
			err = w.retry(err, reqID,
				"encontrar campo de login",
				func() error {
					return pagReceita.ColocarDadosLogin(prestador.Login, prestador.Senha)
				})
			if err != nil {
				return err
			}
		}
		w.escreverErro(true, reqID, "erro genérico ao colocar dados de login: %v", err)
		return fmt.Errorf("erro genérico ao colocar dados de login: %w", err)
	}
	err = pagReceita.ApertarLogin()
	if err != nil {
		switch {
		case errors.Is(err, receita.ErrCaptcha):
			w.escreverMsg(true, reqID, "%v", receita.ErrCaptcha)
			err = pagReceita.BypassCaptcha(false)
			if err != nil {
				if errors.Is(err, receita.ErrNaoEncontrouElemento) {
					w.escreverErro(true, reqID, "%v", err)
					err = w.retry(err, reqID,
						"fazer captcha",
						func() error {
							return pagReceita.BypassCaptcha(true)
						})
				}
				if errors.Is(err, captcha.ErrAPI2Captcha) {
					w.escreverErro(true, reqID, "%v", err)
					return fmt.Errorf("")
				}
			}
		case errors.Is(err, receita.ErrDadosLoginInvalidos):
			w.escreverErro(true, reqID, "%v", receita.ErrDadosLoginInvalidos)
			return receita.ErrDadosLoginInvalidos
		default:
			w.escreverErro(true, reqID, "erro ao efetuar o login: %v", err)
			err = w.retry(err, reqID,
				"efetuar login",
				pagReceita.ApertarLogin)
			if err != nil {
				return err
			}
		}
	}
	w.escreverMsg(true, reqID, "Login feito com sucesso para %s.", prestador.Prestador)

	err = pagReceita.IrParaFormsEmissao()
	if err != nil {
		if !errors.Is(err, receita.ErrNaoCarregaNovaPagina) {
			w.escreverErro(true, reqID, "%v:%v", receita.ErrNaoCarregaNovaPagina, err)
			return fmt.Errorf("%w: %w", receita.ErrNaoCarregaNovaPagina, err)
		}

		err = w.retry(err, reqID,
			"encontrar botão para forms de emissão de NFSe",
			pagReceita.IrParaFormsEmissao)
		if err != nil {
			return err
		}

	}
	w.escreverMsg(false, "", "Forms de emissão de NFSE carregado.")

	for i, nota := range prestador.NotasEmitir {
		w.escreverMsg(true, reqID, "Emitindo nota %d de %d", i+1, len(prestador.NotasEmitir))

		if err := pagReceita.ColocaCnpjEData(nota.Cnpj, nota.Data, false); err != nil {
			if !errors.Is(err, receita.ErrNaoEncontrouElemento) {
				w.escreverErro(true, reqID, "erro genérico ao colocar data/cnpj: %v", err)
				return fmt.Errorf("erro genérico ao colocar data/cnpj: %w", err)
			}

			err = w.retry(err, reqID,
				fmt.Sprintf("colocar data/cnpj na nota %d", i+1),
				func() error {
					return pagReceita.ColocaCnpjEData(nota.Cnpj, nota.Data, true)
				})
			if err != nil {
				return err
			}
		}

		if err := pagReceita.ColocarDadosEEmitirNF(nota.Tomador, nota.Observacao, nota.Valor, false); err != nil {
			if !(errors.Is(err, receita.ErrNaoEncontrouElemento) || errors.Is(err, receita.ErrNaoCarregaNovaPagina) || errors.Is(err, receita.ErrNumeroDeRPSPedidoDuranteEmissao)) {
				w.escreverErro(true, reqID, "erro genérico ao colocar dados da nota %d: %v", i+1, err)
				return fmt.Errorf("erro genérico ao colocar data/cnpj: %w", err)
			}
			if errors.Is(err, receita.ErrNumeroDeRPSPedidoDuranteEmissao) {
				w.escreverErro(true, reqID, "%v", receita.ErrNumeroDeRPSPedidoDuranteEmissao)
				return receita.ErrNumeroDeRPSPedidoDuranteEmissao
			}
			err = w.retry(err, reqID,
				fmt.Sprintf("colocar dados da nota %d", i+1),
				func() error {
					return pagReceita.ColocarDadosEEmitirNF(nota.Tomador, nota.Observacao, nota.Valor, true)
				})
			if err != nil {
				return err
			}
		}

		if err := pagReceita.VoltarParaFormDeNota(); err != nil {
			if !(errors.Is(err, receita.ErrNaoCarregaNovaPagina) || errors.Is(err, receita.ErrNaoEncontrouElemento)) {
				w.escreverErro(true, reqID, "erro genérico ao voltar para form de nota: %v", err)
				return fmt.Errorf("erro genérico ao voltar para form de nota: %w", err)
			}

			err = w.retry(err, reqID,
				"voltar para form de nota",
				pagReceita.VoltarParaFormDeNota)
			if err != nil {
				return err
			}
		}
	}
	w.escreverMsg(true, reqID, "Todas as notas fiscais para %s foram emitidas com sucesso!", prestador.Prestador)
	_ = pagina.Close()
	return nil
}

func (w *Workflow) retry(erro error, reqID, operacao string, fn func() error) error {
	w.ui.Erro(erro)
	w.ui.Msg("Tentando novamente...")

	for i := 1; i <= 3; i++ {
		if err := fn(); err != nil {
			w.ui.Erro("Erro na Tentativa %d", i)
			if i == 3 {
				w.escreverErro(true, reqID, "falha definitiva em %s: %v após 3 tentativas", operacao, err)
				return fmt.Errorf("falha definitiva em %s: %w após 3 tentativas", operacao, err)
			}
			continue
		}
		w.escreverMsg(true, reqID, "Sucesso em %s na tentativa %d", operacao, i)
		w.ui.Sucesso(operacao)
		return nil
	}
	return nil
}

// escreverMsg escreve mensagem e logga o conteúdo com argumentos usando Sprintf. O log somente é feito caso necessarioLog seja true.
//
// ReqID somente é necessário se necessário log
func (w *Workflow) escreverMsg(necessarioLog bool, reqID, msg string, args ...any) {
	w.ui.Msg(msg, args...)
	if necessarioLog {
		w.logger.EscreverMensagemComReqID(reqID, msg, args...)
	}
}

// escreverMsg escreve mensagem e logga o conteúdo com argumentos usando Sprintf. O log somente é feito caso necessarioLog seja true.
//
// ReqID somente é necessário se necessário log
func (w *Workflow) escreverErro(necessarioLog bool, reqID string, msg any, args ...any) {
	w.ui.Erro(msg, args...)
	if necessarioLog {
		msgString := fmt.Sprintf(msg.(string), args...)
		err := errors.New(msgString)
		w.logger.EscreverErroComReqID(reqID, err)
	}
}
