package sistema

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var PathRaiz = mustGetPath()

func mustGetPath() string {
	bin, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("houve um erro grave em buscar o diretório raiz do programa. Verifique as suas permissões de usuário: %w", err))
	}
	return filepath.Dir(bin)
}

// PegarInfoOS pega a informação de qual sistema operacional a aplicação está rodando.
// Exemplo: linux, windows, darwin
//
// Input:
//
// fmt.Printf("%s\n", PegarInfoOS())
//
// Output:
//
// linux
func PegarInfoOS() string {
	return fmt.Sprintf("%s", runtime.GOOS)
}

func IniciarAppComWindows() error {
	return nil
}

func VerificaArgumentos() (string, error) {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if !strings.Contains(arg, "GOOGLE_APPLICATION_CREDENTIALS=") {
			return "", ErrNaoPossuiArgumentoCorreto
		}

		return arg, nil
	}
	return "", ErrNaoPossuiArgumentos
}

func IniciarAppComBash() error {
	cmdCompleto := filepath.Join(PathRaiz, "main")
	AccServiceExport, err := VerificaArgumentos()
	if err != nil {
		return err
	}

	cmd := exec.Command(cmdCompleto)
	cmd.Env = append(os.Environ(), AccServiceExport)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%s: %w", "executar comando bash. verifique se todas as configurações foram feitas corretamente", err)
	}

	return nil
}

// CriarPastaComPathAbsoluto cria uma pasta no sistema com permissões 0700. Não consegue burlar nenhuma permissão de usuário
// Input:
//
// fmt.Println(CriarPastaComPathAbsoluto("/home", "user", "arquivo"))
//
// Output:
//
// /home/user/arquivo <nil>
func CriarPastaComPathAbsoluto(elemen ...string) (string, error) {
	caminho := filepath.Join(elemen...)
	err := os.MkdirAll(caminho, 0700)
	if err != nil {
		return "", fmt.Errorf("%w:%w", ErrMkdir, err)
	}

	return caminho, nil
}

// CriarPastaNaRaiz cria uma pasta na raíz do projeto com um nome definido pelo argumento
func CriarPastaNaRaiz(nome string) (string, error) {
	pathPastaNova := filepath.Join(PathRaiz, nome)

	err := os.MkdirAll(pathPastaNova, 0700)
	if err != nil {
		return "", err
	}

	return pathPastaNova, nil
}

func CriarPastaParaPrestador(prestador string) (string, error) {
	pathNFs, err := CriarPastaNaRaiz("NFs")
	if err != nil {
		return "", err
	}

	pathPrestador := filepath.Join(pathNFs, prestador)

	err = os.MkdirAll(pathPrestador, 0700)
	if err != nil {
		return "", err
	}

	return pathPrestador, nil
}

func LimparPastaNF(path string) error {
	return os.RemoveAll(path)
}
