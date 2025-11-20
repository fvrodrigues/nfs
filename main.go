package main

import (
	"nfse/errs"
	"nfse/logger"
)

func main() {
	arquivoLog, err := logger.CriarArquivoLog()
	if err != nil {
		errs.Formatar("criar pasta /logs ou arquivo de log", err)
	}
	defer arquivoLog.Close()

	logger.EscreverLog(arquivoLog, "Mensagem log de teste..")
	logger.EscreverLog(arquivoLog, "Mensagem log de teste 2.")
	logger.EscreverLog(arquivoLog, "Terminando a execução do programa.")
}
