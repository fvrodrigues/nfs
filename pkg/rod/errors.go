package rod

import "errors"

var (
	ErrNavegacaoAbortada              = errors.New("rod: navegação abortada")
	ErrConfigurarPastaDownloadDefault = errors.New("rod: não foi possível configurar pasta de Download no navegador")
	ErrCriarNavegador                 = errors.New("rod: erro ao criar instãncia de navegador")

	ErrNaoEncontrouElemento = errors.New("rod: falha ao encontrar elemento HTML na página")
)
