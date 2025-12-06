package sheets

import (
	"fmt"
	"strings"
)

type Cliente struct {
	Empresa string

	Login string
	Senha string

	Emitir map[string][]Nota
}

type Nota struct {
	Tomador    string
	Cnpj       string
	Valor      string
	Observacao string

	Emitido bool
	Link    string
}

// TrataValoresDasAbas recebe uma lista de linhas e retorna uma lista de Nota.
func (p *Planilha) TrataValoresDasAbas(linhas [][]any) ([]Nota, error) {
	var notas []Nota

	header := linhas[0]
	indexTomador, indexCnpj, indexValor, indexObservacao, indexEmitido, indexLink := AtribuiValores(header)

	if err := ValidarIndices(indexTomador, indexCnpj, indexValor, indexObservacao, indexEmitido, indexLink); err != nil {
		return nil, err
	}

	for _, linha := range linhas {
		if len(linha) == 0 {
			fmt.Println("Linha vazia, pulando...")
			continue
		}

		var (
			tomador    = ParaStr(linha[indexTomador])
			cnpj       = ParaStr(linha[indexCnpj])
			valor      = ParaValor(linha[indexValor])
			observacao = ParaStr(linha[indexObservacao])
			emitido    = ParaBool(linha[indexEmitido])
			link       = ParaStr(linha[indexLink])
		)

		if !emitido && link != "" {
			linha := Nota{
				Tomador:    tomador,
				Cnpj:       cnpj,
				Valor:      valor,
				Observacao: observacao,
				Emitido:    emitido,
				Link:       link,
			}
			notas = append(notas, linha)
		}
	}
	notas = notas[1:]
	return notas, nil
}

// ParserDados lê os dados de todas as abas e os coloca na struct Nota, além de fazer o parse para o tipo correto de dado.
func (p *Planilha) ParserDados() (map[string][]Nota, error) {
	var notas = map[string][]Nota{}
	abas, err := p.ListarAbas()
	if err != nil {
		return nil, err
	}

	for _, aba := range abas {
		abaToLower := strings.ToLower(aba)
		if strings.Contains(abaToLower, "resposta") {
			fmt.Printf("Pulando aba *%s* pois é de script do forms\n", aba)
			continue
		}
		if strings.Contains(abaToLower, "info") {
			rangeFormatado := fmt.Sprintf("%s!A:Z", aba)
			linhas, err := p.Ler(rangeFormatado)
			if err != nil {
				return nil, err
			}

			for i, linha := range linhas {
				var temCelulaVazia bool

				for _, celula := range linha {
					celula = fmt.Sprintf("%v", celula)
					if celula == "" {
						fmt.Printf("Célula vazia na linha %d, pulando...\n", i+1)
						temCelulaVazia = true
						break
					}
				}
				if temCelulaVazia {
					fmt.Printf("Linha %d vazia, pulando...\n", i+1)
					continue
				}

				// Compara o nome da empresa com o nome da aba.
				empresaToLower := strings.ToLower(linha[0].(string))
				if abaToLower == empresaToLower {

				}
				cliente := Cliente{}
			}
		}

		linhasComValor, err := p.ContarLinhasNaoVazias(aba)
		if err != nil {
			return nil, err
		}
		if linhasComValor <= 1 {
			fmt.Printf("Pulando aba *%s* pois está vazia\n", aba)
			continue
		}

		rangeFormatado := fmt.Sprintf("%s!A:Z", aba)
		linhas, err := p.Ler(rangeFormatado)
		if err != nil {
			return nil, err
		}

		listaNotas, err := p.TrataValoresDasAbas(linhas)
		if err != nil {
			return nil, err
		}

		notas[aba] = append(notas[aba], listaNotas...)
	}

	p.FaltaEmitir = notas
	return notas, nil
}

// AcharIndiceColuna retorna o índice de uma coluna. Recebe como argumento a aba e o nome da coluna.
func AcharIndiceColuna(header []any, str string) int {
	for i, celula := range header {
		celulaString := fmt.Sprintf("%v", celula)
		celulaFormatada := strings.ToLower(strings.TrimSpace(celulaString))
		strFormatada := strings.ToLower(strings.TrimSpace(str))

		if strings.Contains(celulaFormatada, strFormatada) {
			return i
		}
	}
	fmt.Printf("Valor nao encontrado: %s\nHeader usado: %v", str, header)
	return -1
}

func AtribuiValores(header []any) (int, int, int, int, int, int) {
	return AcharIndiceColuna(header, "tomador"), AcharIndiceColuna(header, "cnpj"), AcharIndiceColuna(header, "valor"), AcharIndiceColuna(header, "obs"), AcharIndiceColuna(header, "emiss"), AcharIndiceColuna(header, "link")
}

func LerDadosDeLogin(p *Planilha) (map[string][]Nota, error) {
	return nil, nil
}
