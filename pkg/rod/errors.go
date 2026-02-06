package rod

import "errors"

var (
	ErrNavegacaoAbortada              = errors.New("rod: navegação abortada")
	ErrConfigurarPastaDownloadDefault = errors.New("rod: não foi possível configurar pasta de Download no navegador")
)
