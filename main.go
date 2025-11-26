package main

import (
	"fmt"
	"nfse/pkg/config"
	"nfse/pkg/errs"
	"nfse/pkg/logger"
	"nfse/pkg/receita"
	"nfse/pkg/rod"
	"nfse/pkg/sheets"
	"nfse/pkg/workflow"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go EsperarUmaHora(&wg)

	if err := run(); err != nil {
		panic("Nao deu nada certo")
	}
	wg.Wait()
}

func run() error {
	logger, err := logger.New()
	if err != nil {
		errs.Formatar("criar pasta /logs ou arquivo de log", err)
	}
	defer logger.Fechar()

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("%T, %v", cfg, cfg)
		logger.EscreverMata("ler informações do arquivo .env", err)
	}

	planilha := sheets.NovaPlanilha(logger, cfg.SheetID)
	if err = planilha.NovaConn(); err != nil {
		logger.EscreverMata("criar conexão com a API do Google Sheets", err)
	}

	pagina := rod.CriarNavegador(logger, false)
	defer pagina.Close()

	receita := receita.New(pagina)

	w := workflow.New(cfg, planilha, pagina, receita)
	// return temporário.
	// run() deve retornar sempre um função orquestrante de workflow
	return w.Executar()
}

func EsperarUmaHora(wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Hour)
}
