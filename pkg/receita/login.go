package receita

import (
	"context"
	"errors"
	"fmt"
	"nfse/pkg/rod"
	"strings"
	"time"
)

type Receita struct {
	*rod.Pagina
}

func New(pagina *rod.Pagina) *Receita {
	return &Receita{
		Pagina: pagina,
	}
}

func (r *Receita) AcessarSiteReceita(url string) error {
	if err := r.AcessarSite(url); err != nil {
		return err
	}
	return nil
}

func (r *Receita) ApertarLoginUnico() error {
	wait := r.MustWaitNavigation()
	if err := r.ApertarElemento(".oauth-button"); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("%w: %s", ErrNaoEncontrouElemento, "não foi possível encontrar .oauth-button (login único) na página de login, verifique se o mesmo ainda existe.")
		}
		return fmt.Errorf("%w:%s", err, "erro inesperado achar campo de login único")
	}
	wait()
	return nil
}

func (r *Receita) FazerLogin(cpfCnpj, senha string) error {
	if err := r.EscreverComoHumano("#cpfCnpj", cpfCnpj); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("%w: %s", ErrNaoEncontrouElemento, "não foi possível encontrar o campo #cpfCnpj na página de login, verifique se o mesmo ainda existe.")
		}
		return fmt.Errorf("erro inesperado ao escrever no campo Cpf/Cnpj: %w", err)
	}
	if err := r.EscreverComoHumano("#password", senha); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("%w: %s", ErrNaoEncontrouElemento, "não foi possível encontrar o campo #password na página de login, verifique se o mesmo ainda existe.")
		}
		return fmt.Errorf("erro inesperado ao escrever no campo de senha: %w", err)
	}

	if err := r.ApertarElemento(".btn-entrar"); err != nil {
		return err
	}
	time.Sleep(600 * time.Millisecond)

	url := strings.TrimSpace(strings.ToLower(r.Pagina.MustInfo().URL))
	temErroNaTela, _, _ := r.Pagina.Has(".text-danger")

	// Retorna erro se a url continuar com essa rota ou se o <span> de erro com classe ".text-danger" existir
	if strings.Contains(url, "account/login?ReturnUrl=") || temErroNaTela {
		return fmt.Errorf("%w: %s", ErrDadosLoginInvalidos, "dados de login inválidos")
	}

	if err := r.Timeout(5*time.Second).WaitElementsMoreThan(".ctl00_wpMenuLateral_mnuRotinas_1", 0); err != nil {
		return fmt.Errorf("%w: %s", ErrNaoCarregaNovaPagina, "a página inicial não foi carregada")
	}
	return nil
}

func (r *Receita) Deslogar() error {
	r.PausaHumana(2)
	wait := r.MustWaitNavigation()
	if err := r.ApertarElemento(".oauth-sair"); err != nil {
		return err
	}
	wait()
	return nil
}
