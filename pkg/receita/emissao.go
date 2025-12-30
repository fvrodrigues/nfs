package receita

import (
	"fmt"
	"strings"
	"time"
)

func (r *Receita) ApertarBotaoEmissao() error {
	novoSeletor := "#ctl00_body_ddlApelido" // Elemento que somente existe em nota.aspx

	err := r.AcessarSite("https://nfe.prefeitura.sp.gov.br/contribuinte/nota.aspx")
	if err != nil {
		return err
	}
	r.MustWaitStable()

	if url := r.Pagina.MustInfo().URL; strings.Contains(url, "inicio.aspx") {
		return fmt.Errorf("%w: a página atual ainda contém 'início.aspx ' em sua URL, o que signficia que não foi possível entrar na página de emissão de NFSe", ErrNaoCarregaNovaPagina)
	}

	if err := r.Timeout(10*time.Second).WaitElementsMoreThan(novoSeletor, 0); err != nil {
		return fmt.Errorf("%w: não foi encontrada na página atual o elemento %s, o que significa que a página ainda é a principal", ErrNaoCarregaNovaPagina, novoSeletor)
	}

	return nil
}
