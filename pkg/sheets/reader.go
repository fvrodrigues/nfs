package sheets

import (
	"errors"
	"fmt"
	"nfse/pkg/domain"
	"strings"
	"time"
)

// Prestador é a struct com todos os valores de 'login' e notas para emitir sanitizados e prontos para usar.

func (p *Planilha) ColetarDadosLogin(notas map[string][]domain.Nota) ([]domain.Prestador, error) {
	var prestadores []domain.Prestador
	linhas, err := p.Ler("info!A:C")
	if err != nil {
		return nil, err
	}

	for i, linha := range linhas[1:] {
		var temCelulaVazia bool
		for _, celula := range linha {
			celula = fmt.Sprintf("%v", celula)
			if len(linha) < 3 {
				fmt.Printf("Linha %d não contém informações o suficiente para logar no site. Linha somente possui %d célula(s). Pulando...\n",
					i+2,
					len(linha))
				temCelulaVazia = true
				break
			}
		}
		if temCelulaVazia {
			continue
		}

		nomeEmpresaNaAbaInfo := strings.ToLower(linha[0].(string))
		var achouEmpresa bool

		// Considerando o fato que range em map no golang é aleatório, tive que fazer uma lógica de busca manual
		for nomeEmpresaNaPlanilha, dadosNotas := range notas {
			nomeEmpresaNaPlanilha = strings.ToLower(nomeEmpresaNaPlanilha)
			if nomeEmpresaNaAbaInfo == nomeEmpresaNaPlanilha {
				if dadosNotas == nil || len(dadosNotas) == 0 {
					fmt.Printf("%s não possui notas para emitir. Pulando...\n", nomeEmpresaNaPlanilha)
					achouEmpresa = true
					break
				}
				prestadores = append(prestadores, domain.Prestador{
					Prestador:   linha[0].(string),
					Login:       linha[1].(string),
					Senha:       linha[2].(string),
					NotasEmitir: dadosNotas})
				fmt.Printf("Dados de %s coletados com sucesso\n", nomeEmpresaNaPlanilha)
				achouEmpresa = true
				break
			}
		}

		if !achouEmpresa {
			fmt.Printf("Não foi possível achar empresa %s na planilha. Certifique-se de que o nome na aba 'info' seja exatamente igual ao nome da aba. Pulando...\n", nomeEmpresaNaAbaInfo)
		}

	}
	return prestadores, nil
}

// PegarDadosDeNotasFiscais lê os dados de todas as abas e os coloca na struct Nota, além de fazer o parse para o tipo correto de dado.
func (p *Planilha) PegarDadosDeNotasFiscais() (map[string][]domain.Nota, error) {
	var notas = map[string][]domain.Nota{}
	abas, err := p.ListarAbas()
	if err != nil {
		return nil, err
	}

	for _, aba := range abas {
		abaToLower := strings.ToLower(aba)
		if strings.Contains(abaToLower, "resposta") {
			fmt.Printf("[ . ] Pulando aba *%s* pois é de script do forms\n", aba)
			continue
		}
		if strings.Contains(abaToLower, "logs do") {
			fmt.Printf("[ . ] Pulando aba *%s* pois é a aba de logs\n", aba)
			continue
		}
		if strings.Contains(abaToLower, "info") {
			continue
		}

		rangeFormatado := fmt.Sprintf("%s!A:Z", aba)
		linhas, err := p.Ler(rangeFormatado)
		if err != nil {
			return nil, err
		}
		if len(linhas) <= 1 {
			fmt.Printf("[ ! ] Aba %s somente possui cabeçalho ou está vazia, pulando...\n", aba)
			continue
		}

		listaNotas, err := p.TrataValoresDasAbas(linhas, aba)
		if err != nil {
			if errors.Is(err, ErrFaltaColunaObrigatoria) {
				fmt.Printf("[ ! ] %v\n", err)
				continue
			}
			return nil, err
		}
		linhas = linhas[1:]

		notas[aba] = append(notas[aba], listaNotas...)
		fmt.Printf("Notas de %s coletadas com sucesso\n", aba)
	}
	return notas, nil
}

// TrataValoresDasAbas recebe uma lista de linhas e retorna uma lista de Nota.
func (p *Planilha) TrataValoresDasAbas(linhas [][]any, aba string) ([]domain.Nota, error) {
	var notas []domain.Nota

	header := linhas[0]
	if err := HeaderValido(header, aba); err != nil {
		return nil, err
	}
	indexTomador, indexCnpj, indexValor, indexObservacao, indexEmitido, indexLink, indexEmSP, indexCidade, indexData := AtribuiValores(header)

	if err := ValidarIndices(indexTomador, indexCnpj, indexValor, indexObservacao, indexEmitido, indexLink, indexEmSP, indexCidade); err != nil {
		return nil, err
	}

	for _, linha := range linhas[1:] {
		if len(linha) == 0 {
			continue
		}

		var (
			tomador    = ParaStr(linha[indexTomador])
			cnpj       = ParaStr(linha[indexCnpj])
			valor      = ParaValor(linha[indexValor])
			observacao = ParaStr(linha[indexObservacao])
			emitido    = ParaBool(linha[indexEmitido])
			link       = ParaStr(linha[indexLink])
			data       = ParaStr(linha[indexData])
		)

		data, ok := ParseData(data)
		if !ok {
			data = time.Now().Format("02/01/2006")
		}
		// Como alguns campos podem ficar vazios, se a lógica abaixo não existir essa função vai retornar um monte de notas com valores vazios e true/falses aleatórios.
		if valor == "" || cnpj == "" {
			continue
		}

		if !emitido && link == "" /*&& FoiTotalmenteEmSP(emSP, cidade)*/ {
			linha := domain.Nota{
				Tomador:    tomador,
				Cnpj:       cnpj,
				Valor:      valor,
				Observacao: observacao,
				Emitido:    emitido,
				Link:       link,
				Data:       data,
			}
			notas = append(notas, linha)
		}
	}
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
	return -1
}

func AtribuiValores(header []any) (int, int, int, int, int, int, int, int, int) {
	return AcharIndiceColuna(header, "tomador"),
		AcharIndiceColuna(header, "cnpj"),
		AcharIndiceColuna(header, "valor"),
		AcharIndiceColuna(header, "obs"),
		AcharIndiceColuna(header, "emiss"),
		AcharIndiceColuna(header, "link"),
		AcharIndiceColuna(header, "em SP"),
		AcharIndiceColuna(header, "cidade"),
		AcharIndiceColuna(header, "data")
}

func HeaderValido(header []any, aba string) error {
	colunasObrigatorias := map[string]bool{
		"tomador": false, "cnpj": false, "valor": false, "obs": false, "emissao": false, "link": false, "em sp": false, "cidade": false, "data": false,
	}
	var colunasFaltantes []string

	for _, celula := range header {
		celulaString := fmt.Sprintf("%v", celula)
		celulaFormatada := strings.ToLower(strings.TrimSpace(celulaString))

		switch {
		case strings.Contains(celulaFormatada, "tomador"):
			colunasObrigatorias["tomador"] = true
		case strings.Contains(celulaFormatada, "cnpj"):
			colunasObrigatorias["cnpj"] = true
		case strings.Contains(celulaFormatada, "valor"):
			colunasObrigatorias["valor"] = true
		case strings.Contains(celulaFormatada, "obs"):
			colunasObrigatorias["obs"] = true
		case strings.Contains(celulaFormatada, "emiss"):
			colunasObrigatorias["emissao"] = true
		case strings.Contains(celulaFormatada, "link"):
			colunasObrigatorias["link"] = true
		case strings.Contains(celulaFormatada, "em sp"):
			colunasObrigatorias["em sp"] = true
		case strings.Contains(celulaFormatada, "cidade"):
			colunasObrigatorias["cidade"] = true
		case strings.Contains(celulaFormatada, "data"):
			colunasObrigatorias["data"] = true
		default:
			continue
		}
	}
	for colunaObrigatoria, colunaObrigatoriaExiste := range colunasObrigatorias {
		if !colunaObrigatoriaExiste {
			colunasFaltantes = append(colunasFaltantes, colunaObrigatoria)
		}
	}
	if len(colunasFaltantes) > 0 {
		faltantes := fmt.Sprintf("aba %s não possui seguintes colunas:  %v", aba, colunasFaltantes)
		return fmt.Errorf("%w: %s", ErrFaltaColunaObrigatoria, faltantes) //errors.New(fmt.Sprintf("aba %s faltando colunas: %v", aba, colunasFaltantes))
	}
	return nil
}

func ParseData(dataStr string) (string, bool) {
	data, err := time.Parse("02/01/2006", dataStr)
	if err != nil {
		return "", false
	}
	return data.Format("02/01/2006"), true
}
