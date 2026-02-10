package logger

import (
	"fmt"
	"nfse/pkg/sistema"
	"os"
	"path/filepath"
	"time"
)

type ArquivoLog struct {
	Arquivo *os.File
	Nome    string
}

// New Cria um arquivo de log na pasta /logs com o formato de nome "nfse_dia-mes-ano_hora-minuto-segundo" e retorna o mesmo
func New() (*ArquivoLog, error) {
	// Variável que guarda a hora e datas exatos
	hora := time.Now()

	arquivoNome := fmt.Sprintf("nfse_%v.log", hora.Format("02-01-2006_15-04-05"))

	pastaLogs, err := sistema.CriarPastaNaRaiz("logs")

	caminho := filepath.Join(pastaLogs, arquivoNome)
	arquivo, err := os.OpenFile(caminho, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0500)
	if err != nil {
		return nil, fmt.Errorf("%w:%w", ErrCriarArquivoLog, err)
	}

	return &ArquivoLog{
		arquivo,
		arquivoNome,
	}, nil
}

// format formata a mensagem em formato de log
func (a *ArquivoLog) format(msg string, args ...any) string {
	msg = fmt.Sprintf(msg, args...)
	dataExata := time.Now().Format("02-01-2006 15:04:05.000")
	return fmt.Sprintf("[%s] %s\n", dataExata, msg)
}

func (a *ArquivoLog) formatErro(erro error) string {
	dataExata := time.Now().Format("02-01-2006 15:04:05.000")
	return fmt.Sprintf("[%s] [ ERRO ] %s\n", dataExata, erro.Error())
}

func (a *ArquivoLog) LogInicio(horario string) {
	msg := fmt.Sprintf("[%s] Aplicação iniciada\n", horario)
	if _, err := a.Arquivo.WriteString(msg); err != nil {
		fmt.Println(a.format("[ IMPORTANTE NÃO FATAL ] Logger falhou em iniciar %v:%v\n", ErrEscritaArquivo, err))
	}
}

// EscreverMensagem uma mensagem no arquivo de log.
//
// Caso não consiga escrever, printa o erro
func (a *ArquivoLog) EscreverMensagem(msg string, args ...any) {
	msg = a.format(msg, args...)
	if _, err := a.Arquivo.WriteString(msg); err != nil {
		fmt.Println(a.format("[ IMPORTANTE NÃO FATAL ] Logger falhou em iniciar %v:%v\n", ErrEscritaArquivo, err))
	}
}

// EscreverErro Escreve uma mensagem de erro no arquivo de log.
// Usa a função errs.Formatar da bib interna para formatar a mensagem de erro
//
// Caso não consiga escrever, printa o erro
func (a *ArquivoLog) EscreverErro(erro error) {
	erroFormatado := a.formatErro(erro)
	_, _ = a.Arquivo.WriteString(erroFormatado)
}

// EscreverMensagemComReqID uma mensagem no arquivo de log.
//
// Caso não consiga escrever, printa o erro
func (a *ArquivoLog) EscreverMensagemComReqID(reqID, msg string, args ...any) {
	msg = a.format(msg, args...)
	msg = fmt.Sprintf("[%s] %s", reqID, msg)
	if _, err := a.Arquivo.WriteString(msg); err != nil {
		fmt.Println(a.format("[ IMPORTANTE NÃO FATAL ] Logger falhou em iniciar %v:%v\n", ErrEscritaArquivo, err))
	}
}

// EscreverErroComReqID Escreve uma mensagem de erro no arquivo de log.
// Usa a função errs.Formatar da bib interna para formatar a mensagem de erro
//
// Caso não consiga escrever, printa o erro
func (a *ArquivoLog) EscreverErroComReqID(reqID string, erro error) {
	erroFormatado := a.formatErro(erro)
	erroFormatado = fmt.Sprintf("[%s] %s", reqID, erroFormatado)
	if _, err := a.Arquivo.WriteString(erroFormatado); err != nil {
		fmt.Println(a.format("[ IMPORTANTE NÃO FATAL ] Logger falhou em iniciar %v:%v\n", ErrEscritaArquivo, err))
	}
}

func (a *ArquivoLog) Fechar() {
	_ = a.Arquivo.Close()
}

func (a *ArquivoLog) EncerrarAplicacao() string {
	a.EscreverMensagem("Aplicação encerrada.")
	return "Aplicação encerrada, cheque os logs para mais detalhes."
}
