package errs

import (
	"fmt"
	"log"
)

// Formata a mensagem de erro recebendo como argumento um verbo no infinitivo "action" e o erro.
//
// "Houve o seguinte erro ao <action>: <err>"
//
// "Houve o seguinte erro ao ´se conectar com a API´: <error>"
func Formatar(action string, err error) error {
	return fmt.Errorf("Houve o seguinte erro ao %s: %v", action, err)
}

// Formata a mensagem de erro recebendo como argumento um verbo no infinitivo "action" e o erro, além de printá-la
//
// Além disso mata a aplicação. Somente use em erros que realmente atrapalhariam a execução do programa.
//
// "Houve o seguinte erro ao <action>: <err>"
//
// "Houve o seguinte erro ao ´se conectar com a API´: <error>"
func Mata(action string, err error) {
	log.Fatal(Formatar(action, err))
}
