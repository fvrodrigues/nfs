package config

import (
	"nfse/pkg/sistema"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	SheetID string `env:"SHEET_ID,required"`
	Website string `env:"SITE,required"`
}

// Load carrega o dotenv. Caso ele não esteja na pasta raíz, procura uma antes
func Load(Caminho string) (Config, error) {
	var cfg Config
	caminhoDotEnv := filepath.Join(sistema.PathRaiz, "./.env")
	err := godotenv.Load(caminhoDotEnv)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return cfg, ErrDotEnvNaoEncontrado
		}
		return cfg, err
	}
	if err = env.Parse(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
