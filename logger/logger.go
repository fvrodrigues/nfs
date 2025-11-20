package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Cria um arquivo de log na pasta /logs com o formato de nome "nfse_dia-mes-ano_hora-minuto-segundo" e retorna o mesmo
func CriarArquivoLog() (*os.File, error) {
	// Variável que guarda a hora e datas exatos
	hora := time.Now()

	// Cria o caminho absoluto para a pasta para qualquer OS que esteja executando e nomeia o arquivo
	nomePasta := "logs"
	nomeArquivo := fmt.Sprintf("nfse_%v.log", hora.Format("02-01-2006_15-04-05"))
	caminho := filepath.Join(nomePasta, nomeArquivo)

	if err := os.MkdirAll(nomePasta, 0755); err != nil {
		return nil, err
	}

	arquivo, err := os.OpenFile(caminho, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	EscreverLog(arquivo, "Começando execução do programa.")
	return arquivo, nil
}

// Escreve uma mensagem no arquivo de log e printa a mesma na tela.
//
// Caso não consiga escrever, printa o erro
func EscreverLog(a *os.File, msg string) error {
	// Faz ter certeza ABSOLUTA que a mensagem não tem espaços ou quebras de linha no começo ou final.
	msg = strings.TrimSpace(msg)
	msgFormatada := fmt.Sprintf("[%s] %s\n", time.Now().Format("02-01-2006 15:04:05.000"), msg)

	fmt.Print(msgFormatada)
	if _, err := a.WriteString(msgFormatada); err != nil {
		fmt.Printf("Erro ao escrever no log: %v\n", err)
		return err
	}
	return nil
}
