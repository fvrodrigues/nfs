package receita

import "nfse/pkg/rod"

type Receita struct {
	*rod.Pagina
}

func New(pagina *rod.Pagina) *Receita {
	return &Receita{
		Pagina: pagina,
	}
}

func (r *Receita) Acessar(url string) {
	r.AcessarSite(url)

}

func (r *Receita) Login(cnpj string, senha string) {

}
