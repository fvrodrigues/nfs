package sheets

import (
	"context"
	"fmt"
	"nfse/pkg/logger"

	"google.golang.org/api/sheets/v4"
)

// Planilha é o struct que manterá dados da planilha
type Planilha struct {
	ID string
	*sheets.Service
	Log *logger.ArquivoLog

	Linhas      uint32
	FaltaEmitir map[string][]Nota
}

func NovaPlanilha(l *logger.ArquivoLog, spreadSheetID string) *Planilha {
	return &Planilha{
		ID:  spreadSheetID,
		Log: l,
	}
}

// NewService cria o Service da planilha
func (p *Planilha) NewService() error {
	ctx := context.Background()

	srv, err := sheets.NewService(ctx)
	if err != nil {
		return err
	}

	p.Service = srv
	return nil
}

// Ler popula retorna os valores de um certo range da planilha
//
// [["nome", "idade"], ["papai noel", 2025]]
func (p *Planilha) Ler(celulas string) ([][]any, error) {
	resp, err := p.Service.Spreadsheets.Values.Get(p.ID, celulas).Do()
	if err != nil {
		return nil, err
	}
	return resp.Values, nil
}

// ListarAbas checa as abas/páginas disponíveis da planilha e guarda os nomes um slice de ‘string’.
// As ‘strings’ são formatadas para lowercase
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
