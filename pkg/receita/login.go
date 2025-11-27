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

func (r *Receita) AcessarSiteReceita(url string) error {
	if err := r.AcessarSite(url); err != nil {
		return err
	}
	return nil
}

func (r *Receita) Login(cnpj string, senha string) {

}
