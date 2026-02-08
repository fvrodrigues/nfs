package main

import (
	"fmt"
	"net/http"
	"nfse/pkg/handlers"
	"nfse/pkg/logger"
	"nfse/pkg/sheets"
	"nfse/pkg/ui"
	"nfse/pkg/workflow"
	"sync"
	"time"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	log, err := logger.New()
	if err != nil {
		return fmt.Errorf("%s: %w", "criar pasta /logs ou arquivo de log", err)
	}
	defer log.Fechar()

	planilha := sheets.NovaPlanilha(log, "1eLIdvPoR_X-SW5nKd7fu7mN4i6qp5au55iNYAA8jt9Q")
	if err := planilha.NewService(); err != nil {
		return log.EscreverErro("criar conexão com a API do Google Sheets", err)
	}

	uInterface := ui.NewUI()

	w := workflow.New(log, planilha, uInterface)
	handler := handlers.NewPrestadorHandler(w)

	mux := http.NewServeMux()
	mux.HandleFunc("/prestador", handler.Handle)
	fmt.Println("Servidor rodando na porta 8080")

	server := &http.Server{Addr: ":8080", Handler: mux}

	return server.ListenAndServe()
}

func EsperarParaSempre(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		time.Sleep(time.Hour * 100)
	}
}
