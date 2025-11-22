package sheets

import (
	"context"
	"fmt"
	"nfse/errs"

	"google.golang.org/api/sheets/v4"
)

// Struct que manterá dados da planilha
type Planilha struct {
	ID      string
	Service *sheets.Service

	Linhas   uint32
	Conteudo [][]any
	Clientes []Cliente
}

type Cliente struct {
	Nome    any
	Cnpj    any
	email   any
	emitido any
}

// NovaPlanilha cria um struct com os daods ID e Service da planilha para armazenar os dados necessários para leitura e alteração
//
// Isso porque NewService() não cria uma conexão, somente recebe dados, que são perdidos caso não armazenados
func NovaPlanilha(spreadSheetID string) (*Planilha, error) {
	// Cria o contexto para aplicação
	ctx := context.Background()

	// Método Cloud Native, o google se conecta com a variável de ambiente GOOGLE_APPLICATION_CREDENTIALS
	// A variável deve ser pré-definida com comando export no Linux ou $env no Windows e seu valor deve ser o caminho para o JSON com as infos do Google Service
	srv, err := sheets.NewService(ctx)
	if err != nil {

		return nil, errs.Formatar("se conectar com a API", err)
	}

	return &Planilha{
		ID:      spreadSheetID,
		Service: srv,
	}, nil
}

// Ler popula o campo "Conteudo" do struct, que contém o conteúdo das células
// Recebe como argumento o range das células que serão lidas e popula o campo necessário com seus valores no formato slice de slice de
// any ([][]any). Ou seja, um slice que contém slices dentro que são as linhas, e esses guardam o valor de cada célula
//
// [["nome", "idade"], ["papai noel", 2025]]
func (p *Planilha) Ler(celulas string) ([][]any, error) {
	resp, err := p.Service.Spreadsheets.Values.Get(p.ID, celulas).Do()
	if err != nil {
		return nil, errs.Formatar("ler a planilha.", err)
	}

	return resp.Values, nil
}

// ListarAbas checa as abas/páginas disponíveis da planilha e guarda seus nomes um slice de string.
// Especialmente útil para planilhas que usam várias páginas
func (p *Planilha) ListarAbas() (nomes []string, err error) {
	resp, err := p.Service.Spreadsheets.Get(p.ID).Fields("sheets.properties.title").Do()
	if err != nil {
		return nil, err
	}

	for _, sheet := range resp.Sheets {
		nomes = append(nomes, sheet.Properties.Title)
	}

	return
}

// ContarLinhasNaoVazias verifica a coluna A da aba dada como argumento para ela.
// Retorna um número, que é a quantidade de linhas que iremos trabalhar
// Isso é extremamente
func (p *Planilha) ContarLinhasNaoVazias(aba string) (abas uint32, err error) {
	intervalo := fmt.Sprintf("%s!A:A", aba)
	resp, err := p.Service.Spreadsheets.Values.Get(p.ID, intervalo).Do()
	if err != nil {
		return 0, err
	}

	if len(resp.Values) == 0 {
		return 0, nil
	}

	for _, linha := range resp.Values {
		if len(linha) > 0 && linha[0] != "" {
			abas++
		}
	}

	return
}
