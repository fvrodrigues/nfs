package domain

import "errors"

var (
	ErrFaltaValorObrigatorio = errors.New("dados de requisição: dado(s) obrigatório(s) faltando no corpo da requisição")
)
