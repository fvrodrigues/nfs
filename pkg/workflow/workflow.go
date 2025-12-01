package workflow

import (
	"nfse/pkg/config"
	"nfse/pkg/logger"
	"nfse/pkg/receita"
	"nfse/pkg/rod"
	"nfse/pkg/sheets"
)

type Workflow struct {
	logger   *logger.ArquivoLog
	cfg      config.Config
	planilha *sheets.Planilha
	pagina   *rod.Pagina
	receita  *receita.Receita
}

func New(logger *logger.ArquivoLog, cfg config.Config, planilha *sheets.Planilha, pagina *rod.Pagina, receita *receita.Receita) *Workflow {
	return &Workflow{
		logger:   logger,
		planilha: planilha,
		pagina:   pagina,
		receita:  receita,
		cfg:      cfg,
	}
}

func (w *Workflow) Executar() error {
	defer w.logger.EncerrarAplicacao()
	// w.receita.AcessarSiteReceita(w.cfg.Website)
	w.planilha.ParserDados()
	return nil
}
