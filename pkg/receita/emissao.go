package receita

import (
	"time"

	"github.com/go-rod/rod/lib/proto"
)

func (r *Receita) IrParaFormsEmissao() error {
	if err := r.EsperarEstabilidade(1*time.Minute, 400*time.Millisecond); err != nil {
		return r.wrapErroLoad(err)
	}
	r.PausaHumana(2)
	err := r.AcessarSite(siteBase() + "/contribuinte/nota.aspx")
	if err != nil {
		return err
	}

	if err := r.EsperarEstabilidade(1*time.Minute, 400*time.Millisecond); err != nil {
		return r.wrapErroLoad(err)
	}

	return nil
}

func (r *Receita) ColocaCnpjEData(cnpj, data string, apagar bool) error {
	if err := r.DigitarTecladoComoHumano("#ctl00_body_tbCPFCNPJTomador", cnpj, apagar); err != nil {
		return r.wrapErrorApertarElemento(err, "#ctl00_body_tbCPFCNPJTomador", "digitar cnpj")
	}

	data = r.DataParaDigitarComoHumano(data)
	if err := r.DigitarTecladoComoHumano("#ctl00_body_txtEmitidoEm", data, apagar); err != nil {
		return r.wrapErrorApertarElemento(err, "#ctl00_body_txtEmitidoEm", "digitar data")
	}

	if err := r.ApertarElemento("#ctl00_body_btAvancar"); err != nil {
		return r.wrapErrorApertarElemento(err, "ctl00_body_btAvancar", "avançar página")
	}

	// Evita o 'bug' das datas. erroData somente vai existir e fazer a condição ser true se o 'bug' ocorrer
	if _, err := r.Timeout(400 * time.Millisecond).Element("#ctl00_body_lblMensagem"); err == nil {
		if err := r.ApertarElemento("#ctl00_body_rbVersao1"); err != nil {
			return r.wrapErrorApertarElemento(err, "#ctl00_body_rbVersao1", "selecionar Versão 1 - ISS apenas")
		}

		if err := r.ApertarElemento("#ctl00_body_btAvancar"); err != nil {
			return r.wrapErrorApertarElemento(err, "ctl00_body_btAvancar", "avançar página")
		}
	}

	if err := r.EsperarEstabilidade(1*time.Minute, 400*time.Millisecond); err != nil {
		return r.wrapErroLoad(err)
	}
	return nil
}

func (r *Receita) ColocarDadosEEmitirNF(nome, obs, valor string, apagar bool) error {
	if err := r.EsperarEstabilidade(1*time.Minute, 400*time.Millisecond); err != nil {
		return r.wrapErroLoad(err)
	}

	if err := r.DigitarTecladoComoHumano("#ctl00_body_tbRazaoSocial", nome, apagar); err != nil {
		return r.wrapErrorApertarElemento(err, "#ctl00_body_tbRazaoSocial", "digitar nome do tomador")
	}

	if err := r.DigitarTecladoComoHumano("#ctl00_body_tbDiscriminacao", obs, apagar); err != nil {
		return r.wrapErrorApertarElemento(err, "ctl00_body_tbDiscriminacao", "colocar discriminação do serviço")
	}

	if err := r.DigitarTecladoComoHumano("#ctl00_body_tbValor", valor, apagar); err != nil {
		return r.wrapErrorApertarElemento(err, "#ctl00_body_tbValor", "digitar valor da NF")
	}

	if err := r.ApertarDialogEmitir(); err != nil {
		return r.wrapErrorApertarElemento(err, "#ctl00_body_btEmitir", "emitir nota")
	}
	if err := r.EsperarEstabilidade(1*time.Minute, 400*time.Millisecond); err != nil {
		return r.wrapErroLoad(err)
	}

	if _, err := r.Timeout(500 * time.Millisecond).Search("Número do RPS não preenchido"); err == nil {
		return ErrNumeroDeRPSPedidoDuranteEmissao
	}

	if err := r.ApertarElemento("#btDownload"); err != nil {
		return r.wrapErrorApertarElemento(err, "#btDownload", "download de NF")
	}
	return nil
}

func (r *Receita) VoltarParaFormDeNota() error {
	r.PausaHumana(2)

	wait := r.Page.WaitNavigation(proto.PageLifecycleEventNameLoad)
	if err := r.ApertarElemento("#ctl00_cphBase_btVoltar"); err != nil {
		return r.wrapErrorApertarElemento(err, "#ctl00_cphBase_btVoltar", "voltar para form de cpf e data")
	}
	wait()
	return nil
}

func (r *Receita) ApertarDialogEmitir() error {
	r.PausaHumana(2)

	// Espera window.confirm do JS que aparece ao apertar botão de emitir e confirma emissão
	wait, handle := r.MustHandleDialog()
	go func() {
		wait()
		handle(true, "")
	}()
	if err := r.ApertarElemento("#ctl00_body_btEmitir"); err != nil {
		return err
	}

	return nil
}
