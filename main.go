package main

import (
	"fmt"
	"nfse/errs"
	"nfse/logger"
	"nfse/rod"
	"nfse/sheets"
	"sync"
	"time"
)

func main() {
	// Para teste, faz a aplicação esperar uma hora antes de fechar
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done() // avisa quando terminar
		fmt.Println("Routine iniciada.")
		time.Sleep(time.Hour) // espera 1 hora
		fmt.Println("Routine finalizada.")
	}()
	// ------------------------------------------------------------

	logger, err := logger.CriarArquivoLog()
	if err != nil {
		errs.Formatar("criar pasta /logs ou arquivo de log", err)
	}
	defer logger.Fechar()

	planilha := sheets.NovaPlanilha(logger, "fvbaFPp4wjlKt448lpMZv2Yze2qCTLVpjp4w")
	if err = planilha.NovaConn(); err != nil {
		logger.EscreverLogMata("criar conexão com a API do Google Sheets", err)
	}

	pag := rod.CriarNavegador(logger, false)
	defer pag.Pagina.Close()

	pag.AcessarSite("https://nfe.prefeitura.sp.gov.br/login.aspx")
	pag.ApertarElemento(".oauth-name")
	time.Sleep(4 * time.Second)

	fmt.Printf("Execução do programada terminada. Cheque os logs em %v\n", logger.Caminho)

	wg.Wait() // espera a goroutine terminar
	fmt.Println("Programa encerrado.")
}
