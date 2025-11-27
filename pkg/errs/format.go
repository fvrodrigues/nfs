package errs

import (
	"fmt"
	"log"
	"os"
)

// Formata a mensagem de erro recebendo como argumento um verbo no infinitivo "action" e o erro.
//
// "Houve o seguinte erro ao <action>: <err>"
//
// "Houve o seguinte erro ao ´se conectar com a API´: <error>"
func Formatar(action string, err error) error {
	return fmt.Errorf("Erro ao %s: %v", action, err)
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

func Encerrar(err error, caminho string) {
	fmt.Printf("O programa encerrou pelo seguinte motivo:\n %v\nVeja os logs em %s", err, caminho)
	os.Exit(1)
}
