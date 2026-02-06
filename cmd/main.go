package main

import (
	"fmt"
	"net/http"
	"nfse/pkg/config"
	"nfse/pkg/handlers"
	"nfse/pkg/logger"
	"nfse/pkg/sheets"
	"nfse/pkg/ui"
	"nfse/pkg/workflow"
	"sync"
	"time"
)

func main() {

	// Apagar depois que tudo tiver pronto
	var wg sync.WaitGroup
	wg.Add(1)
	go EsperarUmaHora(&wg)
	// --------

	if err := run(); err != nil {
		panic(err)
	}

	// --------
	wg.Wait()
	// --------
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

	ui := ui.NewUI()

	w := workflow.New(logger, cfg, planilha, ui)
	handler := handlers.NewPrestadorHandler(w)

	mux := http.NewServeMux()
	mux.HandleFunc("/prestador", handler.Handle)
	fmt.Println("Servidor rodando na porta 8080")

	server := &http.Server{Addr: ":8080", Handler: mux}

	return server.ListenAndServe()
}

func EsperarUmaHora(wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Minute * 30)
}
