package sheets

import (
	"fmt"
	"strings"
)

// Cliente é a struct com todos os valores de 'login' e notas para emitir sanitizados e prontos para usar.
type Cliente struct {
	Empresa string

	Login string
	Senha string

	NotasEmitir []Nota
}

type Nota struct {
	Tomador    string
	Cnpj       string
	Valor      string
	Observacao string

	Emitido bool
	Link    string
}

func (p *Planilha) ColetarDadosLogin(notas map[string][]Nota) ([]Cliente, error) {
	//DEBUG ------------
	fmt.Println("---------------")
	fmt.Println("---------------")
	fmt.Println("---------------")
	for k, v := range notas {
		if len(v) == 0 {
			fmt.Printf("%s está vazio:\n", k)
		}
	}
	fmt.Println("---------------")
	fmt.Println("---------------")
	fmt.Println("---------------")
	//FIM DO DEBUG ------------

	var clientes []Cliente
	linhas, err := p.Ler("info!A:D")
	if err != nil {
		return nil, err
	}

	for i, linha := range linhas[1:] {
		var temCelulaVazia bool
		for _, celula := range linha {
			celula = fmt.Sprintf("%v", celula)
			if celula == "" {
				fmt.Printf("\n------------------------------\nCélula vazia na linha %d, pulando...\n", i+2)
				temCelulaVazia = true
				break
			}
			if len(linha) < 3 {
				fmt.Printf("\n------------------------------\nLinha %d não contém informações o suficiente para logar no site.\nEnquanto que precisamos de 3 informações, somente temos %d\nPulando...\n",
					i+2,
					len(linha))
				temCelulaVazia = true
				break
			}
		}
		if temCelulaVazia {
			continue
		}

		nomeEmpresaAbaInfo := strings.ToLower(linha[0].(string))

		// Considerando o fato que range em map no golang é aleatório, tive que fazer uma lógica de busca manual
		for nomeEmpresaPlanilha, dadosNotas := range notas {
			if nomeEmpresaAbaInfo == strings.ToLower(nomeEmpresaPlanilha) {
				if dadosNotas == nil || len(dadosNotas) == 0 {
					fmt.Printf("\n------------------------------\n%s não possui notas para emitir. Pulando...\n------------------------------\n", nomeEmpresaPlanilha)
					break
				}
				fmt.Printf("\n------------------------------\nLogin de %s encontrado.\n", nomeEmpresaPlanilha)
				clientes = append(clientes, Cliente{
					Empresa:     linha[0].(string),
					Login:       linha[1].(string),
					Senha:       linha[2].(string),
					NotasEmitir: dadosNotas})
				fmt.Printf("Dados coletados para %s\n------------------------------\n", nomeEmpresaPlanilha)
				break
			}
		}
	}
	return clientes, nil
}

// PegarDadosDeNotasFiscais lê os dados de todas as abas e os coloca na struct Nota, além de fazer o parse para o tipo correto de dado.
func (p *Planilha) PegarDadosDeNotasFiscais() (map[string][]Nota, error) {
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
			continue
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

	return notas, nil
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

		if !emitido || link == "" {
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
