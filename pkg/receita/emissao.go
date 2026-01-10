package receita

import (
	"fmt"
	"strings"
	"time"
)

func (r *Receita) ApertarBotaoEmissao() error {
	const seletorOpcaoEmitir = "#ctl00_body_ddlApelido" // Elemento que somente existe em nota.aspx

	_ = r.ApertarElemento(seletorOpcaoEmitir)

	err := r.AcessarSite("https://nfe.prefeitura.sp.gov.br/contribuinte/nota.aspx")
	if err != nil {
		return err
	}
	r.MustWaitStable()

	if url := r.Pagina.MustInfo().URL; !strings.Contains(url, "nota.aspx") {
		return fmt.Errorf("%w: a página atual não contém 'nota.aspx' em sua URL, o que signficia que não foi possível entrar na página de emissão de NFSe", ErrNaoCarregaNovaPagina)
	}

	if err := r.Timeout(5*time.Second).WaitElementsMoreThan(seletorOpcaoEmitir, 0); err != nil {
		return fmt.Errorf("%w: não foi encontrada na página atual o elemento %s, o que significa que a página ainda é a principal", ErrNaoCarregaNovaPagina, seletorOpcaoEmitir)
	}

	return nil
}

func (r *Receita) ColocaCnpjEData(cnpj, data string) error {
	err := r.DigitarTecladoComoHumano("#ctl00_body_tbCPFCNPJTomador", cnpj)
	if err != nil {
		if strings.Contains(err.Error(), "deadline exceeded") {
			return fmt.Errorf("%s:%w", "não foi possível encontrar #ctl00_body_tbCPFCNPJTomador (input de cnpj) na página de emissão, verifique se o mesmo ainda existe.", ErrNaoEncontrouElemento)
		}
		return fmt.Errorf("%s:%w", "erro inesperado ao tentar digitar cnpj do tomador", err)
	}

	data = r.DataParaDigitar(data)
	err = r.DigitarTecladoComoHumano("#ctl00_body_txtEmitidoEm", data)
	if err != nil {
		if strings.Contains(err.Error(), "deadline exceeded") {
			return fmt.Errorf("%s:%w", "não foi possível encontrar #ctl00_body_txtEmitidoEm (input de data) na página de emissão, verifique se o mesmo ainda existe.", ErrNaoEncontrouElemento)
		}
		return fmt.Errorf("%s:%w", "erro inesperado ao tentar data prestação do serviço", err)
	}

	err = r.ApertarElemento("#ctl00_body_btAvancar")
	if err != nil {
		if strings.Contains(err.Error(), "deadline exceeded") {
			return fmt.Errorf("%s:%w", "não foi possível encontrar #ctl00_body_btAvancar (botão de avançar página) na página de emissão, verifique se o mesmo ainda existe.", ErrNaoEncontrouElemento)
		}
		return fmt.Errorf("%s:%w", "erro inesperado ao tentar avançar", err)
	}
	r.MustWaitStable()
	return nil
}

func (r *Receita) DataParaDigitar(data string) string {
	slData := strings.Split(data, "/")
	return fmt.Sprintf("%s%s%s", slData[0], slData[1], slData[2])
}
