package captcha

import (
	"fmt"
	"os"

	api2captcha "github.com/2captcha/2captcha-go"
)

type Captcha struct {
	Client        *api2captcha.Client
	NormalCaptcha api2captcha.Normal
}

func New() *Captcha {
	return &Captcha{
		Client: api2captcha.NewClient(os.Getenv("API_KEY")),
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
