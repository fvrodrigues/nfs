package config

import "errors"

var (
	ErrDotEnvNaoEncontrado = errors.New(".env: não foi possível encontrar o arquivo .env na raiz do projeto")
)
