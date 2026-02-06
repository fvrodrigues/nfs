package sistema

import "errors"

var (
	ErrNaoPossuiArgumentos       = errors.New("sistema: execução do comando feita sem argumentos")
	ErrNaoPossuiArgumentoCorreto = errors.New(`sistema: o argumento entregue está inválido, o mesmo deve ser "GOOGLE_APPLICATION_CREDENTIALS=<caminho/para/servicce_account.json>"`)
)
