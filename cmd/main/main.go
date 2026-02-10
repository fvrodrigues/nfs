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
		return fmt.Errorf("%w", err)
	}
	horaAplicativoIniciado := time.Now().Format("02-01-2006 15:04:05.000")
	log.LogInicio(horaAplicativoIniciado)
	defer log.Fechar()

	planilha := sheets.NovaPlanilha(log, "1eLIdvPoR_X-SW5nKd7fu7mN4i6qp5au55iNYAA8jt9Q")
	if err := planilha.NewService(); err != nil {
		log.EscreverErro(err)
		return err
	}

	uInterface := ui.NewUI()

	w := workflow.New(log, planilha, uInterface)
	handler := handlers.NewPrestadorHandler(w)

	mux := http.NewServeMux()
	mux.HandleFunc("/prestador", handler.HandlePost)
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
