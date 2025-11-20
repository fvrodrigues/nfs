package sheets

import (
	"context"
	"nfse/errs"
	"time"

	"google.golang.org/api/sheets/v4"
)

// Struct que manterá dados da planilha
type Planilha struct {
	ID      string
	Service *sheets.Service
}

func NovaPlanilha(spreadSheetId string) (*Planilha, error) {
	// Cria o contexto para aplicação
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Método Cloud Native, o google se conecta com a variável de ambiente GOOGLE_APPLICATION_CREDENTIALS
	// A variável deve ser pré-definida com comando export no Linux ou $env no Windows e seu valor deve ser o caminho para o JSON com as infos do Google Service
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, errs.Formatar("se conectar com a API", err)
	}

	return &Planilha{
		ID:      spreadSheetId,
		Service: srv,
	}, nil
}

func (p *Planilha) Ler(celulas string) ([][]interface{}, error) {
	resp, err := p.Service.Spreadsheets.Values.Get(p.ID, celulas).Do()
	if err != nil {
		return nil, errs.Formatar("ler a planilha.", err)
	}
	return resp.Values, nil
}
