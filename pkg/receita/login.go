package receita

import (
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

func (r *Receita) AcessarSiteReceita() error {
	err := r.AcessarSite("https://nfe.prefeitura.sp.gov.br/login.aspx")
	if err == nil {
		return r.EsperarEstabilidade(1*time.Minute, 400*time.Millisecond)
	}

	if strings.Contains(err.Error(), "ERR_ABORTED") {
		return fmt.Errorf("%w: sessão abortada ao tentar acessar website da receita. Tente novamente", ErrSessaoAbortada)
	}
	return fmt.Errorf("erro genérico ao acessar website da receita: %w", err)
}

func (r *Receita) EncontrarBotaoLoginUnico() error {
	err := r.EsperarEstabilidade(40*time.Second, 400*time.Millisecond)
	if err != nil {
		return ErrNaoCarregaNovaPagina
	}
	_, err = r.RetornaElemento(".oauth-button", 500*time.Millisecond)
	if err != nil {
		return r.wrapErrorApertarElemento(err, ".oauth-button", "encontrar botão de login único")
	}
	return nil
}

func (r *Receita) ApertarLoginUnico() error {
	err := r.EsperarEstabilidade(40*time.Second, 400*time.Millisecond)
	if err != nil {
		return ErrNaoCarregaNovaPagina
	}
	err = r.ApertarElemento(".oauth-button")
	if err != nil {
		return fmt.Errorf("%s:%w", "erro inesperado achar campo de login único", err)
	}

	return r.EsperarEstabilidade(40*time.Second, 400*time.Millisecond)
}

func (r *Receita) ColocarDadosLogin(cpfCnpj, senha string) error {
	err := r.EsperarEstabilidade(40*time.Second, 400*time.Millisecond)
	if err != nil {
		return ErrNaoCarregaNovaPagina
	}
	r.PausaHumana(2)

	if err := r.EscreverComoHumano("#cpfCnpj", cpfCnpj); err != nil {
		return r.wrapErrorApertarElemento(err, "#cpfCnpj", "escrever no campo de cpf/cnpj")
	}
	if err := r.EscreverComoHumano("#password", senha); err != nil {
		return r.wrapErrorApertarElemento(err, "#password", "escrever no campo de senha")
	}
	return nil
}

func (r *Receita) ApertarLogin() error {
	r.PausaHumana(2)

	err := r.EsperarEstabilidade(40*time.Second, 400*time.Millisecond)
	if err != nil {
		return ErrNaoCarregaNovaPagina
	}
	r.PausaHumana(2)

	if err := r.ApertarElemento(".btn-entrar"); err != nil {
		return r.wrapErrorApertarElemento(err, ".btn-entrar", "fazer login")
	}

	if existe, _, _ := r.Timeout(400 * time.Millisecond).Has(".text-danger"); existe {
		return ErrDadosLoginInvalidos
	}
	if err := r.EsperarEstabilidade(1*time.Minute, 400*time.Millisecond); err != nil {
		return r.wrapErroLoad(err)
	}
	return nil
}

func (r *Receita) Deslogar() error {
	r.PausaHumana(2)

	if err := r.ApertarElemento("#ctl00_HeaderV2_oauthLogout_Sair"); err != nil {
		fmt.Printf("Não foi possível encontrar o elemento #ctl00_HeaderV2_oauthLogout_Sair para deslogar, verifique se o mesmo ainda existe. Tentando deslogar manualmente, o que pode alertar da automação\n")
	}

	url := strings.TrimSpace(strings.ToLower(r.Pagina.MustInfo().URL))
	if !strings.Contains(url, "notadomilhao.sf.") {
		fmt.Printf("Saindo manualmente com JS.\n")

		_, err := r.Eval(`__doPostBack('ctl00$HeaderV2$oauthLogout$Sair','')`)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrNaoEncontrouElemento, "não foi possível sair manualmente usando o JavaScript da página.")
		}
		err = r.EsperarEstabilidade(1*time.Minute, 400*time.Millisecond)
		if err != nil {
			return err
		}
	}

	//Verifica novamente se não deu certo. Possivelmente uma das piores e mais redundantes linhas de código que já escrevi na vida, mas o site do governo não me ajuda.
	if !strings.Contains(url, "notadomilhao.sf.") {
		return fmt.Errorf("%w: %s", ErrNaoCarregaNovaPagina, "não foi possível deslogar")
	}
	return nil
}
