package sheets

import (
	"fmt"
	"strings"
)

func ValidarIndices(indicesParaValidar ...int) error {
	for _, indice := range indicesParaValidar {
		if indice < 0 {
			return fmt.Errorf("A coluna de um dos valores necessários para emissão está faltando")
		}
	}
	return nil
}

func ParaBool(valor any) bool {
	str := fmt.Sprintf("%v", valor)
	str = strings.ToLower(strings.TrimSpace(str))

	// Protege caso a planilha envie o resultado de outros modos
	switch str {
	case "sim", "true", "1", "ok":
		return true
	}
	return false
}

func ParaStr(valor any) string {
	str := fmt.Sprintf("%v", valor)
	return strings.ToLower(strings.TrimSpace(str))
}

func ParaValor(valor any) string {
	str := fmt.Sprintf("%v", valor)
	str = strings.ReplaceAll(str, "R$", "")
	str = strings.TrimSpace(str)

	return str
}
