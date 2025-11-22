package main

import (
	"fmt"
	"nfse/errs"
	"nfse/logger"
)

func main() {
	arquivoLog, err := logger.CriarArquivoLog()
	if err != nil {
		errs.Formatar("criar pasta /logs ou arquivo de log", err)
	}

	fmt.Printf("Execução do programada terminada. Cheque os logs em %v\n", arquivoLog.Caminho)
}
