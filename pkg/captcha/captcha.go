package captcha

import (
	"fmt"

	api2captcha "github.com/2captcha/2captcha-go"
)

type Captcha struct {
	Client        *api2captcha.Client
	NormalCaptcha api2captcha.Normal
}

func New() *Captcha {
	return &Captcha{
		Client: api2captcha.NewClient("3f7122b69478a5e4c19fcd92a6b6c583"),
	}
}

func (c *Captcha) Resolve(base64 string) (string, string, error) {
	c.NormalCaptcha = api2captcha.Normal{
		Base64: base64,
	}
	code, ID, err := c.Client.Solve(c.NormalCaptcha.ToRequest())
	if err != nil {
		return "", ID, fmt.Errorf("ID requisição [%s] %w:%w", ID, ErrAPI2Captcha, err)
	}
	return code, ID, nil
}
