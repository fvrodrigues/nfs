package receita

import "errors"

var (
	ErrDadosLoginInvalidos  = errors.New("login: falha ao logar no website da receita")
	ErrNaoEncontrouElemento = errors.New("login: falha ao encontrar elemento HTML na página")

	ErrNaoCarregaNovaPagina = errors.New("emissão: não foi possível ir para nova página")
)
