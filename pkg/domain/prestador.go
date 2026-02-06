package domain

import "fmt"

type Prestador struct {
	Prestador string `json:"prestador" validate:"required,min=3,max=100"`

	Login string `json:"login" validate:"min=3,max=100"`
	Senha string `json:"senha" validate:"min=3,max=100"`

	NotasEmitir []Nota `json:"notas" validate:"required,dive,required"`
}

type Nota struct {
	Tomador    string `json:"tomador" validate:"required,min=3,max=100"`
	Cnpj       string `json:"cnpj" validate:"required,cpf,cnpj"`
	Valor      string `json:"valor" validate:"required,numeric,min=0.01"`
	Observacao string `json:"observacao" validate:"required,min=3,max=100"`
	Data       string `json:"data" validate:"required,date_format=01/02/2006"`

	Emitido bool   `json:"emitido"`
	Link    string `json:"link"`
}

func (p *Prestador) ValidaDadosCorpoReq() error {
	switch {
	case p.Prestador == "":
		return fmt.Errorf("%w: %s", ErrFaltaValorObrigatorio, "nome do prestador de serviços")
	case p.Login == "":
		return fmt.Errorf("%w: %s", ErrFaltaValorObrigatorio, "login site do governo")
	case p.Senha == "":
		return fmt.Errorf("%w: %s", ErrFaltaValorObrigatorio, "senha site do governo")
	case p.NotasEmitir == nil || len(p.NotasEmitir) == 0:
		return fmt.Errorf("%w: %s", ErrFaltaValorObrigatorio, "nota(s) para emitir")
	}

	for _, nota := range p.NotasEmitir {
		switch {
		case nota.Tomador == "":
			return fmt.Errorf("%w: %s", ErrFaltaValorObrigatorio, "nome do tomador")
		case nota.Cnpj == "":
			return fmt.Errorf("%w: %s", ErrFaltaValorObrigatorio, "cpf ou cnpj do tomador")
		case nota.Valor == "":
			return fmt.Errorf("%w: %s", ErrFaltaValorObrigatorio, "valor do serviços")
		case nota.Observacao == "":
			return fmt.Errorf("%w: %s", ErrFaltaValorObrigatorio, "discriminação de serviço")
		case nota.Data == "":
			return fmt.Errorf("%w: %s", ErrFaltaValorObrigatorio, "data prestação de serviço")
		}
	}

	return nil
}
