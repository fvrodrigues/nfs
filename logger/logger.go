package logger

import (
	"fmt"
	"nfse/errs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ArquivoLog struct {
	File    *os.File
	Nome    string
	Caminho string
}

// Cria um arquivo de log na pasta /logs com o formato de nome "nfse_dia-mes-ano_hora-minuto-segundo" e retorna o mesmo
func CriarArquivoLog() (*ArquivoLog, error) {
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

	return &ArquivoLog{
		arquivo,
		nomeArquivo,
		caminho,
	}, nil
}

// Escreve uma mensagem no arquivo de log.
//
// Caso não consiga escrever, printa o erro
func (a *ArquivoLog) EscreverLog(msg string) error {
	// Faz ter certeza ABSOLUTA que a mensagem não tem espaços ou quebras de linha no começo ou final.
	msg = strings.TrimSpace(msg)
	dataExata := time.Now().Format("02-01-2006 15:04:05.000")
	msgFormatada := fmt.Sprintf("[%s] %s\n", dataExata, msg)

	if _, err := a.File.WriteString(msgFormatada); err != nil {
		fmt.Printf("Erro ao escrever no log: %v\n", err)
		return err
	}
	return nil
}

// Escreve uma mensagem de erro no arquivo de log.
// Usa a função errs.Formatar da bib interna para formatar a mensagem de erro
//
// Caso não consiga escrever, printa o erro
func (a *ArquivoLog) EscreverLogErro(action string, erro error) error {
	erro = errs.Formatar(action, erro)
	dataExata := time.Now().Format("02-01-2006 15:04:05.000")
	msgFormatada := fmt.Sprintf("[%s] %v\n", dataExata, erro)

	if _, err := a.File.WriteString(msgFormatada); err != nil {
		fmt.Printf("Erro ao escrever no log: %v\n", err)
		return err
	}
	return nil
}

// Escreve uma mensagem de erro no arquivo de log, além de matar o programa.
// Usa a função errs.Formatar da bib interna para formatar a mensagem de erro
//
// Caso não consiga escrever, printa o erro
func (a *ArquivoLog) EscreverLogMata(action string, erro error) error {
	dataExata := time.Now().Format("02-01-2006 15:04:05.000")
	msgFormatada := fmt.Sprintf("[%s] %v\n", dataExata, erro)

	if _, err := a.File.WriteString(msgFormatada); err != nil {
		fmt.Printf("Erro ao escrever no log: %v\n", err)
		return err
	}
	errs.Mata(action, erro)
	return nil
}

// NaoAchaElemento escreve uma mensagem no log e mata o programa.
//
// Usado especialmente quando o programa falhar em encontrar um elemento HTML
func (a *ArquivoLog) NaoAchaElemento(HTMLElement string, err error) error {
	return a.EscreverLogMata(fmt.Sprintf("tentar encontrar o elemento %s", HTMLElement), err)
}

func (l *ArquivoLog) Fechar() {
	l.File.Close()
}
