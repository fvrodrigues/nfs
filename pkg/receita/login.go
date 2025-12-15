package receita

import (
	"fmt"
	"nfse/pkg/rod"
	"strings"
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
		return err
	}
	wait()
	return nil
}

func (r *Receita) FazerLogin(cpfCnpj, senha string) error {
	if err := r.EscreverComoHumano("#cpfCnpj", cpfCnpj); err != nil {
		return err
	}
	if err := r.EscreverComoHumano("#password", senha); err != nil {
		return err
	}

	wait := r.MustWaitNavigation()
	if err := r.ApertarElemento(".btn-entrar"); err != nil {
		return err
	}
	wait()

	url := strings.TrimSpace(strings.ToLower(r.Pagina.MustInfo().URL))
	temErroNaTela, _, _ := r.Pagina.Has(".text-danger")

	// Retorna erro se a url continuar com essa rota ou se o <span> de erro com classe ".text-danger" existir
	if strings.Contains(url, "account/login?ReturnUrl=") || temErroNaTela {
		return fmt.Errorf("dados de login inválidos")
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
