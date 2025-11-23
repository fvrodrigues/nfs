package sheets

import (
	"context"
	"fmt"
	"nfse/logger"

	"google.golang.org/api/sheets/v4"
)

// Struct que manterá dados da planilha
type Planilha struct {
	ID string
	*sheets.Service
	Log *logger.ArquivoLog

	Linhas   uint32
	Conteudo [][]any
	Clientes []Cliente
}

func NovaPlanilha(l *logger.ArquivoLog, spreadSheetID string) *Planilha {
	return &Planilha{
		ID:  spreadSheetID,
		Log: l,
	}
}

type Cliente struct {
	Nome    any
	Cnpj    any
	email   any
	emitido any
}

// NovaConn cria o *sheets.Service da planilha
//
// Isso porque NewService() não cria uma conexão, somente recebe dados, que são perdidos caso não armazenados
func (p *Planilha) NovaConn() error {
	// Cria o contexto para aplicação
	ctx := context.Background()

	// Método Cloud Native, o google se conecta com a variável de ambiente GOOGLE_APPLICATION_CREDENTIALS
	// A variável deve ser pré-definida com comando export no Linux ou $env no Windows e seu valor deve ser o caminho para o JSON com as infos do Google Service
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return err
	}

	p.Service = srv
	return nil
}

// Ler popula o campo "Conteudo" do struct, que contém o conteúdo das células
// Recebe como argumento o range das células que serão lidas e popula o campo necessário com seus valores no formato slice de slice de
// any ([][]any). Ou seja, um slice que contém slices dentro que são as linhas, e esses guardam o valor de cada célula
//
// [["nome", "idade"], ["papai noel", 2025]]
func (p *Planilha) Ler(celulas string) ([][]any, error) {
	resp, err := p.Service.Spreadsheets.Values.Get(p.ID, celulas).Do()
	if err != nil {
		return nil, err
	}
	return resp.Values, nil
}

// ListarAbas checa as abas/páginas disponíveis da planilha e guarda seus nomes um slice de string.
// Especialmente útil para planilhas que usam várias páginas
func (p *Planilha) ListarAbas() (nomes []string, err error) {
	resp, err := p.Service.Spreadsheets.Get(p.ID).Fields("sheets.properties.title").Do()
	if err != nil {
		p.Log.EscreverLogErro("listar abas da planilha", err)
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
		p.Log.EscreverLogErro("contar as linhas não vazias", err)
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
