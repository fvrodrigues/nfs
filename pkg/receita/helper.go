package receita

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod/lib/input"
)

func (r *Receita) wrapErrorApertarElemento(err error, seletor, acao string) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "deadline exceeded") {
		return fmt.Errorf("%w: não foi possível encontrar %s (%s), verifique se o elemento mudou", ErrNaoEncontrouElemento, seletor, acao)
	}
	return fmt.Errorf("erro inesperado ao %s: %w", acao, err)
}

func (r *Receita) wrapErrorApertarTecla(err error, tecla, acao string) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "deadline exceeded") {
		return fmt.Errorf("%w: não foi possível apertar tecla %s para %s", ErrNaoEncontrouElemento, tecla, acao)
	}
	return fmt.Errorf("erro inesperado ao %s: %w", acao, err)
}

func (r *Receita) wrapErroEncontrarElemento(err error, el, acao string) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "deadline exceeded") {
		return fmt.Errorf("%w: não foi possível encontrar %s (%s), verifique se o elemento mudou", ErrNaoCarregaNovaPagina, el, acao)
	}
	return fmt.Errorf("erro inesperado ao %s: %w", acao, err)
}

func (r *Receita) wrapErroLoad(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "deadline exceeded") {
		return ErrConexao
	}
	return fmt.Errorf("erro inesperado ao obter resposta do servidor: %w", err)
}

func (r *Receita) DataParaDigitarComoHumano(data string) string {
	slData := strings.Split(data, "/")
	return fmt.Sprintf("%s%s%s", slData[0], slData[1], slData[2])
}

func (r *Receita) PressionarTecla(key input.Key) {
	r.KeyActions().Press(key)
	r.KeyActions().Release(key)
}

// EsperarEstabilidade espera tanto o carregamento da página quanto o carregamento dos elementos
//
// Usa WaitStable para ver se a página fica estável pela quantidade de ms após carregamento. Se em caso não fique em 20seg, retorna erro
func (r *Receita) EsperarEstabilidade(sec, ms time.Duration) error {
	err := r.Timeout(sec).WaitLoad()
	if err != nil {
		return err
	}
	return r.Timeout(sec).WaitStable(ms)
}

func DataEhValida(dataStr string) bool {
	_, err := time.Parse("02/01/2006", dataStr)
	if err != nil {
		return false
	}
	return true
}
