package logger

import "errors"

var (
	ErrLeituraArquivo  = errors.New("logger: falha ao ler arquivo")
	ErrEscritaArquivo  = errors.New("logger: falha ao escrever arquivo")
	ErrCriarArquivoLog = errors.New("logger: falha ao criar arquivo de log")
	ErrCriarPastaLogs  = errors.New("logger: falha ao criar pasta de logs")
)
