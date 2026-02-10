package ui

import "fmt"

// UI somente será usad no workflow
type UI struct {
	Erro    func(any, ...any)
	Sucesso func(string, ...any)
	Msg     func(string, ...any)
}

func NewUI() *UI {
	return &UI{
		Erro: func(msgErro any, args ...any) {
			msg := fmt.Sprintf("%v", msgErro)
			mensagem := fmt.Sprintf(msg, args...)
			fmt.Printf("[ ERRO ] %s\n", mensagem)
		},
		Sucesso: func(mensagem string, args ...any) {
			mensagem = fmt.Sprintf(mensagem, args...)
			fmt.Printf("[ SUCESSO ] %s\n", mensagem)
		},
		Msg: func(mensagem string, args ...any) {
			mensagem = fmt.Sprintf(mensagem, args...)
			fmt.Printf("[ . ] %s\n", mensagem)
		},
	}
}
