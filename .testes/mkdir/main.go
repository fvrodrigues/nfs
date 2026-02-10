package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

func main() {
	a := NewReqID()
	fmt.Println(a)
}

func CriarPastaComPathAbsoluto(elemen ...string) (string, error) {
	caminho := filepath.Join(elemen...)
	err := os.MkdirAll(caminho, 0700)
	if err != nil {
		return "", err
	}

	return caminho, nil
}

func DoisRetornos() (string, error) {
	return "um", nil
}

func Format(msg string, args ...any) string {
	msg = fmt.Sprintf(msg, args...)
	dataExata := time.Now().Format("02-01-2006 15:04:05.000")
	return fmt.Sprintf("[%s] %s\n", dataExata, msg)
}

func escreve(mensagem string, args ...any) {
	fmt.Printf("[ SUCESSO ] " + mensagem + "\n")
}

func NewReqID() string {
	// 64 bits aleatórios => ~13 chars em base36
	max := new(big.Int).Lsh(big.NewInt(1), 64)
	n, _ := rand.Int(rand.Reader, max)
	return n.Text(36)
}
