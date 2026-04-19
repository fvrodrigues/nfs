package receita

import (
	"fmt"
	"nfse/pkg/captcha"
	"nfse/pkg/rod"
	"os"
	"strings"
	"time"
)

// defaultSiteBase é a URL base do portal da Receita Municipal de São Paulo.
// Em tempo de execução, pode ser sobrescrita pela variável de ambiente SITE,
// o que é usado principalmente em testes para apontar para um portal "mock".
const defaultSiteBase = "https://nfe.prefeitura.sp.gov.br"

// Seletores da página de Login Único. A SP trocou o formulário de
// IDs dedicados (#cpfCnpj, #password, .btn-entrar) para um formulário
// ASP.NET padrão com campos por `name="Username"`/`name="Password"` e
// botões de submit sem classe específica. Os nomes permanecem estáveis
// porque são exigidos pelo backend.
const (
	loginCPFCNPJSelector = `input[name="Username"]`
	loginSenhaSelector   = `input[name="Password"]`
	loginEntrarSelector  = `form button[type="submit"]`
)

// siteBase retorna a URL base do portal da Receita, respeitando a variável
// de ambiente SITE quando ela estiver definida.
func siteBase() string {
	if v := strings.TrimSpace(os.Getenv("SITE")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return defaultSiteBase
}

type Receita struct {
	Captcha *captcha.Captcha
	*rod.Pagina
}

func New(pagina *rod.Pagina, captcha *captcha.Captcha) *Receita {
	return &Receita{
		Captcha: captcha,
		Pagina:  pagina,
	}
}

func (r *Receita) AcessarSiteReceita() error {
	err := r.AcessarSite(siteBase() + "/login.aspx")
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
	_, err = r.RetornaElemento(".oauth-button", 5*time.Millisecond)
	if err != nil {
		return r.wrapErrorApertarElemento(err, ".oauth-button", "encontrar botão de login único")
	}
	return nil
}

func (r *Receita) ApertarLoginUnico() error {
	err := r.EsperarEstabilidade(120*time.Second, 400*time.Millisecond)
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
	// Espera a página ficar estável
	err := r.EsperarEstabilidade(40*time.Second, 400*time.Millisecond)
	if err != nil {
		return ErrNaoCarregaNovaPagina
	}

	r.PausaHumana(2)

	// Espera o campo de CPF/CNPJ ser visível antes de escrever.
	// A página de Login Único (pmspauth.prefeitura.sp.gov.br) usa
	// `input[name="Username"]` / `input[name="Password"]` em vez de
	// IDs dedicados para o CPF/CNPJ e a senha.
	elCpf, err := r.Timeout(10 * time.Second).Element(loginCPFCNPJSelector)
	if err != nil {
		return r.wrapErrorApertarElemento(err, loginCPFCNPJSelector, "escrever no campo de cpf/cnpj")
	}
	elCpf.MustWaitVisible() // Garante que o campo está visível

	if err := r.EscreverComoHumano(loginCPFCNPJSelector, cpfCnpj); err != nil {
		return r.wrapErrorApertarElemento(err, loginCPFCNPJSelector, "escrever no campo de cpf/cnpj")
	}

	// Espera o campo de senha ser visível antes de escrever
	elSenha, err := r.Timeout(10 * time.Second).Element(loginSenhaSelector)
	if err != nil {
		return r.wrapErrorApertarElemento(err, loginSenhaSelector, "escrever no campo de senha")
	}
	elSenha.MustWaitVisible() // Garante que o campo está visível

	if err := r.EscreverComoHumano(loginSenhaSelector, senha); err != nil {
		return r.wrapErrorApertarElemento(err, loginSenhaSelector, "escrever no campo de senha")
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

	if err := r.ApertarElemento(loginEntrarSelector); err != nil {
		return r.wrapErrorApertarElemento(err, loginEntrarSelector, "fazer login")
	}

	if existe, _, _ := r.Timeout(400 * time.Millisecond).Has(".text-danger"); existe {
		return ErrDadosLoginInvalidos
	}

	if err := r.EsperarEstabilidade(1*time.Minute, 400*time.Millisecond); err != nil {
		return r.wrapErroLoad(err)
	}
	if r.TemCaptcha() {
		return ErrCaptcha
	}
	return nil
}

func (r *Receita) BypassCaptcha(deveApagar bool) error {
	_ = r.EsperarEstabilidade(40*time.Second, 400*time.Millisecond)
	r.PausaHumana(2)

	code, _, err := r.PegaValoresCaptcha()
	if err != nil {
		return err // Erro já formatado
	}

	if err := r.DigitarTecladoComoHumano("#ans", code, deveApagar); err != nil {
		return r.wrapErroEncontrarElemento(err, "#ans", "input para digitar captcha")
	}
	if err := r.ApertarElemento("#jar"); err != nil {
		return r.wrapErrorApertarElemento(err, "#jar", "botão para enviar captcha")
	}

	return nil
}

func (r *Receita) PegaValoresCaptcha() (string, string, error) {
	base64, err := r.RetornaBase64Captcha() //ErrNaoEncontrouElemento
	if err != nil {
		return "", "", fmt.Errorf("erro ao pegar base64 da imagem captcha: %w", err)
	}
	stringB64 := fmt.Sprintf("%v", base64)
	return r.Captcha.Resolve(stringB64)
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
