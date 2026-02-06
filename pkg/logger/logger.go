package logger

import (
	"fmt"
	"nfse/pkg/sistema"
	"os"
	"path/filepath"
	"time"
)

type ArquivoLog struct {
	File    *os.File
	Nome    string
	Caminho string
}

// New Cria um arquivo de log na pasta /logs com o formato de nome "nfse_dia-mes-ano_hora-minuto-segundo" e retorna o mesmo
func New() (*ArquivoLog, error) {
	// Variável que guarda a hora e datas exatos
	hora := time.Now()

	arquivoNome := fmt.Sprintf("nfse_%v.log", hora.Format("02-01-2006_15-04-05"))

	pastaLogs, err := sistema.CriarPasta("logs")
	if err != nil {
		return nil, err
	}

	caminho := filepath.Join(pastaLogs, arquivoNome)
	arquivo, err := os.OpenFile(caminho, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &ArquivoLog{
		arquivo,
		arquivoNome,
		caminho,
	}, nil
}

// Format formata a mensagem em formato de log
func (a *ArquivoLog) Format(msg any) (msgFormatada string) {
	dataExata := time.Now().Format("02-01-2006 15:04:05.000")
	msgFormatada = fmt.Sprintf("[%s] %s\n", dataExata, msg)
	return
}

// Escrever uma mensagem no arquivo de log.
//
// Caso não consiga escrever, printa o erro
func (a *ArquivoLog) Escrever(msg string) error {
	msg = a.Format(msg)
	if _, err := a.File.WriteString(msg); err != nil {
		fmt.Printf("Erro ao escrever no log: %v\n", err)
		return err
	}
	return nil
}

// EscreverErro Escreve uma mensagem de erro no arquivo de log.
// Usa a função errs.Formatar da bib interna para formatar a mensagem de erro
//
// Caso não consiga escrever, printa o erro
func (a *ArquivoLog) EscreverErro(action string, erro error) error {
	erroFormatado := a.Format(erro)

	if _, err := a.File.WriteString(erroFormatado); err != nil {
		fmt.Printf("Erro ao escrever no log: %v\n", err)
		return err
	}
	return erro
}

// NaoAchaElemento escreve uma mensagem no log e mata o programa.
// API do Google Sheets
// Usado especialmente quando o programa falhar em encontrar um elemento HTML
func (a *ArquivoLog) NaoAchaElemento(HTMLElement string, err error) error {
	return a.EscreverErro(fmt.Sprintf("tentar encontrar o elemento %s", HTMLElement), err)
}

func (a *ArquivoLog) Fechar() {
	_ = a.File.Close()
}

func (a *ArquivoLog) EncerrarAplicacao() {
	_ = a.Escrever("Aplicação encerrada.")
	fmt.Println("Aplicação encerrada, cheque os logs em ", a.Caminho)
}
