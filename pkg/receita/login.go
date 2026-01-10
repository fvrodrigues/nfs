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
	err := r.AcessarSite(url)
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "ERR_ABORTED") {
		return fmt.Errorf("%w: sessão abortada ao tentar acessar %s. Tente novamente", ErrSessaoAbortada, url)
	}
	return fmt.Errorf("erro genérico ao acessar site %s: %w", url, err)
}

func (r *Receita) ApertarLoginUnico() error {
	err := r.ApertarElemento(".oauth-button")
	if err == nil {
		r.MustWaitStable()
		return nil
	}

	if strings.Contains(err.Error(), "deadline exceeded") {
		return fmt.Errorf("%w: %s", ErrNaoEncontrouElemento, "não foi possível encontrar .oauth-button (login único) na página de login, verifique se o mesmo ainda existe.")
	}
	return fmt.Errorf("%w:%s", err, "erro inesperado achar campo de login único")
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
		if strings.Contains(err.Error(), "deadline exceeded") {
			return fmt.Errorf("%w: %s", ErrNaoEncontrouElemento, "não foi possível encontrar .btn-entrar (botão 'Entrar') na página de login, verifique se o mesmo ainda existe.")
		}
		return fmt.Errorf("%w:%s", err, "erro inesperado ao tentar fazer login")
	}
	time.Sleep(600 * time.Millisecond)
	r.MustWaitStable()

	url := strings.TrimSpace(strings.ToLower(r.Pagina.MustInfo().URL))
	temErroNaTela, _, _ := r.Pagina.Has(".text-danger")

	// Retorna erro se a url continuar com essa rota ou se o <span> de erro com classe ".text-danger" existir
	if strings.Contains(url, "account/login?returnurl=") || temErroNaTela {
		return fmt.Errorf("%w: %s", ErrDadosLoginInvalidos, "dados de login inválidos")
	}

	if err := r.Timeout(5*time.Second).WaitElementsMoreThan(".ctl00_wpMenuLateral_mnuRotinas_1", 0); err != nil {
		return fmt.Errorf("%w: %s", ErrNaoCarregaNovaPagina, "a página inicial não foi carregada")
	}
	return nil
}

func (r *Receita) Deslogar() error {
	r.PausaHumana(2)

	if err := r.ApertarElemento(".oauth-sair"); err != nil {
		fmt.Printf("Não foi possível encontrar o elemento .ouath-sair para deslogar, verifique se o mesmo ainda existe. Tentando deslogar manualmente, o que pode alertar da automação\n")
	}
	_ = r.WaitStable(2 * time.Second)

	time.Sleep(5 * time.Second)
	url := strings.TrimSpace(strings.ToLower(r.Pagina.MustInfo().URL))
	if !strings.Contains(url, "notadomilhao.sf.") {
		fmt.Printf("Saindo manualmente com JS.\n")

		_, err := r.Eval(`__doPostBack('ctl00$HeaderV2$oauthLogout$Sair','')`)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrNaoEncontrouElemento, "não foi possível sair manualmente usando o JavaScript da página.")
		}
		_ = r.WaitStable(2 * time.Second)
	}

	//Verifica novamente se não deu certo. Possivelmente uma das piores e mais redundantes linhas de código que já escrevi na vida, mas o site do governo não me ajuda.
	if !strings.Contains(url, "notadomilhao.sf.") {
		return fmt.Errorf("%w: %s", ErrNaoCarregaNovaPagina, "não foi possível deslogar")
	}
	return nil
}
