package workflow

import (
	"fmt"
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

	dadosNotasFiscais, err := w.planilha.PegarDadosDeNotasFiscais()
	if err != nil {
		return err
	}
	dadosDosClientes, err := w.planilha.ColetarDadosLogin(dadosNotasFiscais)
	if err != nil {
		return err
	}

	// Tirar
	for i, _ := range dadosDosClientes {
		fmt.Printf("-------------\nDados:\nNome: %s\nLogin: %s\nSenha: %s\n\nDados emissão: %v\n---------------------\n",
			dadosDosClientes[i].Empresa,
			dadosDosClientes[i].Login,
			dadosDosClientes[i].Senha,
			dadosDosClientes[i].NotasEmitir)
	}
	// Tirar

	return nil
}
