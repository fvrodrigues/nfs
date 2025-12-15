package workflow

import (
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

	w.ui.Msg("Pegando dados das notas fiscais na planilha...")
	dadosNotasFiscais, err := w.planilha.PegarDadosDeNotasFiscais()
	if err != nil {
		return err
	}

	w.ui.Msg("Coletando dados de login e vendo as notas que precisam ser emitidas...")
	clientes, err := w.planilha.ColetarDadosLogin(dadosNotasFiscais)
	if err != nil {
		return err
	}
	if clientes == nil {
		w.ui.Sucesso("Não há notas fiscais pendentes para emitir. Encerrando aplicação")
		return nil
	}

	for _, cliente := range clientes {
		if err := w.receita.AcessarSiteReceita(w.cfg.Website); err != nil {
			return err
		}
		w.ui.MsgComArg("Acessado: %s", w.cfg.Website)

		if err := w.receita.ApertarLoginUnico(); err != nil {
			return err
		}

		if err := w.receita.FazerLogin(cliente.Login, cliente.Senha); err != nil {
			if err.Error() == "dados de login inválidos" {
				w.ui.Erro("%v, pulando usuário...", err)
				continue
			}
			return err
		}
		w.ui.MsgComArg("Login feito para %v", cliente.Empresa)

		if err := w.receita.Deslogar(); err != nil {
			return err
		}

	}
	return nil

}
