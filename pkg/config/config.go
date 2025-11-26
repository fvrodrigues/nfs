package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	SheetID string `env:"SHEET_ID,required"`
	Website string `env:"SITE,required"`
}

func Load() (cfg Config, err error) {
	if err = godotenv.Load(); err != nil {
		return
	}
	if err = env.Parse(&cfg); err != nil {
		return
	}
	return
}
