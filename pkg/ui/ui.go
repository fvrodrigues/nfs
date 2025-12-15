package ui

import "fmt"

// UI somente será usad no workflow
type UI struct {
	Erro          func(string, any)
	Sucesso       func(string)
	SucessoComArg func(string, any)
	Msg           func(string)
	MsgComArg     func(string, any)
}

func NewUI() *UI {
	return &UI{
		Erro: func(mensagem string, arg any) {
			fmt.Printf("[ FALHA ] "+mensagem+"\n", arg)
		},
		Sucesso: func(mensagem string) {
			fmt.Printf("[ SUCESSO ] " + mensagem + "\n")
		},
		SucessoComArg: func(mensagem string, arg any) {
			fmt.Printf("[ SUCESSO ] "+mensagem+"\n", arg)
		},
		Msg: func(mensagem string) {
			fmt.Printf("[ WORKFLOW ] " + mensagem + "\n")
		},
		MsgComArg: func(mensagem string, arg any) {
			fmt.Printf("[ WORKFLOW ] "+mensagem+"\n", arg)
		},
	}
}
