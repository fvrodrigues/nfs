package sheets

import (
	"context"
	"fmt"
	"nfse/pkg/logger"
	"strings"
	"time"

	"google.golang.org/api/sheets/v4"
)

// Struct que manterá dados da planilha
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

type Nota struct {
	Tomador    any //string
	Cnpj       any //string
	Valor      any //float64
	Observacao any //string

	Emitido bool
	Link    string
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

// Lê os dados de todas as abas e os coloca na struct Nota no campo Emitir, além de fazer o parse para o tipo correto de dado.
// O campo emitir guarda detalhes de todas as notas que ainda precisam ser emitidas. Começa a lê de aba!A2-F<número de linhas>
//
//	Essa função não tem a capacidade de ler nomes no header, ela somente assume que a planilha estará totalmente na ordem
func (p *Planilha) ParserDados() (map[string][]Nota, error) {
	notas := make(map[string][]Nota)
	abas, err := p.ListarAbas()
	if err != nil {
		return nil, err
	}

	for _, aba := range abas {
		abaToLower := strings.ToLower(aba)
		if strings.Contains(abaToLower, "resposta") || strings.Contains(abaToLower, "info") {
			fmt.Printf("Pulando aba *%s* pois é de forms ou info\n", aba)
			continue
		}

		linhasComValor, err := p.ContarLinhasNaoVazias(aba)
		if err != nil {
			return nil, err
		}
		if linhasComValor == 0 {
			fmt.Printf("Pulando aba *%s* pois está vazia\n", aba)
			continue
		}

		rangeFormatado := fmt.Sprintf("%s!A2:F%d", aba, linhasComValor)
		linhas, err := p.Ler(rangeFormatado)
		if err != nil {
			return nil, err
		}
		fmt.Println(aba)

		for _, linha := range linhas {
			if len(linha) == 0 {
				// O problema dessa parte do código é que quando ele encontra uma linha vazia, ele não escaneia a última linha. Devo arrumar um modo de corrigir isso
				fmt.Printf("Linha vazia na aba %s. Pulando\n", aba)
				continue
			}
			novaLinha := Nota{
				Tomador:    linha[0],
				Cnpj:       linha[1],
				Valor:      linha[2],
				Observacao: linha[3],
				Emitido:    true,
				Link:       "ads",
			}
			fmt.Println(novaLinha)
			time.Sleep(100 * time.Millisecond)

			notas[aba] = append(notas[aba], novaLinha)
		}
		fmt.Println("----------------------")
		time.Sleep(1000 * time.Millisecond)

	}
	fmt.Printf("===========================================\n\n===========================================\n")
	func(m map[string][]Nota) {
		for key, lista := range m {
			fmt.Printf("%q:\n", key)
			for _, item := range lista {
				fmt.Printf("    %+v\n", item)
			}
			fmt.Println("-----------")
		}
	}(notas)
	p.FaltaEmitir = notas
	return nil, nil
}

// ListarAbas checa as abas/páginas disponíveis da planilha e guarda seus nomes um slice de string.
// As strings são formatadas para lowercase
// Especialmente útil para planilhas que usam várias páginas
func (p *Planilha) ListarAbas() (nomes []string, err error) {
	resp, err := p.Service.Spreadsheets.Get(p.ID).Fields("sheets.properties.title").Do()
	if err != nil {
		return nil, err
	}

	for _, sheet := range resp.Sheets {
		nomeFormatado := strings.ToLower(sheet.Properties.Title)
		nomes = append(nomes, nomeFormatado)
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
