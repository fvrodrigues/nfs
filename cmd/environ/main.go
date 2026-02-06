package main

import (
	"errors"
	"fmt"
	"nfse/pkg/sistema"
	"strings"
)

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}

func run() error {
	OS := sistema.PegarInfoOS()
	switch {
	case strings.Contains(OS, "linux"):
		err := sistema.IniciarAppComBash()
		if err != nil {
			return fmt.Errorf("não foi possível iniciar aplicação principal: %w", err)
		}
		return nil
	case strings.Contains(OS, "windows"):
		err := sistema.IniciarAppComWindows()
		if err != nil {
			return err
		}
	default:
		return errors.New("sistema operacional selecionado não é suportado")
	}
	return nil
}
