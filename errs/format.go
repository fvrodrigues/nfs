package errs

import (
	"fmt"
)

// Formata a mensagem de erro recebendo como argumento um verbo no infinitivo "action" e o erro.
//
// "Houve o seguinte erro ao <action>: <err>"
//
// "Houve o seguinte erro ao ´se conectar com a API´: <error>"
func Formatar(action string, err error) error {
	return fmt.Errorf("Houve o seguinte erro ao %s: %v", action, err)
}
