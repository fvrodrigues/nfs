package main

import (
	"fmt"
	"net/http"
	"nfse/cmd/math"
	"nfse/pkg/captcha"
	"nfse/pkg/handlers"
	"nfse/pkg/logger"
	"nfse/pkg/ui"
	"nfse/pkg/workflow"
	"time"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	porta := math.Port()

	log, err := logger.New()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	horaAplicativoIniciado := time.Now().Format("02-01-2006 15:04:05.000")
	log.LogInicio(horaAplicativoIniciado)
	defer log.Fechar()

	solver := captcha.New()
	uInterface := ui.NewUI()

	w := workflow.New(log, solver, uInterface)
	handler := handlers.NewPrestadorHandler(w)

	mux := http.NewServeMux()
	mux.HandleFunc("/prestador", handler.HandlePost)
	fmt.Printf("Servidor rodando na porta %s\n", porta)

	server := &http.Server{Addr: porta, Handler: mux}

	return server.ListenAndServe()
}
