package captcha

import "fmt"

var (
	ErrAPI2Captcha     = fmt.Errorf("2captcha: erro interno na API")
	ErrCaptchaInvalido = fmt.Errorf("2captcha: código de captcha recebido inválido")
)
