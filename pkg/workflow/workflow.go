package workflow

import (
	"fmt"
	"nfse/pkg/config"
	"nfse/pkg/receita"
	"nfse/pkg/rod"
	"nfse/pkg/sheets"
)

type Workflow struct {
	cfg      config.Config
	planilha *sheets.Planilha
	pagina   *rod.Pagina
	receita  *receita.Receita
}

func New(cfg config.Config, planilha *sheets.Planilha, pagina *rod.Pagina, receita *receita.Receita) *Workflow {
	return &Workflow{
		planilha: planilha,
		pagina:   pagina,
		receita:  receita,
		cfg:      cfg,
	}
}

func (w *Workflow) Executar() error {
	w.receita.Acessar(w.cfg.Website)
	return nil
}
