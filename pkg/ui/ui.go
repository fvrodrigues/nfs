package ui

import "fmt"

// UI somente será usad no workflow
type UI struct {
	Erro           func(error)
	Sucesso        func(string)
	SucessoComArg  func(string, any)
	Msg            func(string)
	MsgComArg      func(string, any)
	Workflow       func(string)
	WorkflowComArg func(string, any)
}

func NewUI() *UI {
	return &UI{
		Erro: func(err error) {
			fmt.Printf("[ FALHA ] %v\n", err)
		},
		Sucesso: func(mensagem string) {
			fmt.Printf("[ SUCESSO ] " + mensagem + "\n")
		},
		SucessoComArg: func(mensagem string, arg any) {
			fmt.Printf("[ SUCESSO ] "+mensagem+"\n", arg)
		},
		Msg: func(mensagem string) {
			fmt.Printf("[ . ] " + mensagem + "\n")
		},
		MsgComArg: func(mensagem string, arg any) {
			fmt.Printf("[ . ] "+mensagem+"\n", arg)
		},
		Workflow: func(mensagem string) {
			fmt.Printf("[ WORKFLOW ] " + mensagem + "\n")
		},
		WorkflowComArg: func(mensagem string, arg any) {
			fmt.Printf("[ WORKFLOW ] "+mensagem+"\n", arg)
		},
	}
}
