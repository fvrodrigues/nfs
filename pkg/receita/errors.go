package receita

import "errors"

var (
	ErrDadosLoginInvalidos  = errors.New("login: dados de login inválidos")
	ErrNaoEncontrouElemento = errors.New("login: falha ao encontrar elemento HTML na página")
	ErrSessaoAbortada       = errors.New("login: sessão abortada")

	ErrNaoCarregaNovaPagina            = errors.New("emissão: não foi possível ir para nova página")
	ErrNumeroDeRPSPedidoDuranteEmissao = errors.New("emissão: RPS pedido durante emissão. Verifique a data inserida e tente novamente")

	ErrConexao = errors.New("página: não foi possível estabelecer comunicação com o serviço no momento. A solicitação foi interrompida devido a instabilidade de rede ou indisponibilidade temporária do website desejado")
)
