package main

import (
	"fmt"
	"nfse/pkg/config"
	"nfse/pkg/logger"
	"nfse/pkg/receita"
	"nfse/pkg/rod"
	"nfse/pkg/sheets"
	"nfse/pkg/ui"
	"nfse/pkg/workflow"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go EsperarUmaHora(&wg)

	if err := run(); err != nil {
		panic(err)
	}
	wg.Wait()
}

func run() error {
	logger, err := logger.New()
	if err != nil {
		return fmt.Errorf("%w: %s", "criar pasta /logs ou arquivo de log", err)
	}
	defer logger.Fechar()

	cfg, err := config.Load()
	if err != nil {
		return logger.EscreverErro("ler informações do arquivo .env", err)
	}

	planilha := sheets.NovaPlanilha(logger, cfg.SheetID)
	if err := planilha.NewService(); err != nil {
		return logger.EscreverErro("criar conexão com a API do Google Sheets", err)
	}

	pagina, err := rod.CriarNavegador(logger, false)
	if err != nil {
		return logger.EscreverErro("criar instância de navegador", err)
	}
	defer pagina.Close()

	rec := receita.New(pagina)

	ui := ui.NewUI()

	w := workflow.New(logger, cfg, planilha, pagina, rec, ui)
	return w.Executar()
}

func EsperarUmaHora(wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Minute * 30)
}
