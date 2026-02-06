package rod

import (
	"fmt"
	"math/rand"
	"nfse/pkg/logger"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
)

type Pagina struct {
	Log       *logger.ArquivoLog
	Navegador *rod.Browser
	// Mesmo com embedded field, o nome desse campo é "Page"
	*rod.Page
}

// CriarNavegador cria uma instância do navegador configurado para a automação e inicia o logger.
// Se true, navegador virá headless
func CriarNavegador(log *logger.ArquivoLog, headless bool) (*Pagina, error) {
	l := launcher.New().Headless(headless).Devtools(false)
	l.Set("disable-gpu")
	l.Set("no-sandbox")
	l.Set("disable-dev-shm-usage")
	l.Set("excludeSwitches", "enable-automation")
	l.Set("useAutomationExtension", "false")
	l.Set("start-maximized", "true")

	url, err := l.Launch()
	if err != nil {
		return nil, err
	}

	browser := rod.New().ControlURL(url).NoDefaultDevice().MustConnect()
	page := stealth.MustPage(browser)

	return &Pagina{
		Log:       log,
		Navegador: browser,
		Page:      page,
	}, nil
}

// AcessarSite Com um navegador e página já instanciados
func (p *Pagina) AcessarSite(url string) error {

	if err := p.Navigate(url); err != nil {
		return fmt.Errorf("erro ao acessar site %s: %w", url, err)
	}
	p.MustWaitStable()
	return nil
}

// ApertarElemento usa a própria bib rod para fazer movimentos humanos e clicar em um botao ou elemento especificado
// Recebe HTMLElement como argumento.
func (p *Pagina) ApertarElemento(HTMLElement string) error {
	botao, err := p.RetornaElemento(HTMLElement, 2)
	if err != nil {
		return fmt.Errorf("erro ao encontrar elemento %s: %w\n", HTMLElement, err)
	}

	botao.MustHover()
	p.PausaHumana(2)
	p.Mouse.MustDown(proto.InputMouseButtonLeft)
	p.PausaHumana(0)
	p.Mouse.MustUp(proto.InputMouseButtonLeft)

	return nil
}

// ApertarElementoDefinido Assim como ApertarElemento clica num elemento com movimento humano exceto que recebe o próprio elemento como argumento para não procurar na página. Útil quando o elemento já está declarado.
func (p *Pagina) ApertarElementoDefinido(elemento *rod.Element) {
	elemento.MustHover()
	p.PausaHumana(2)
	p.Mouse.MustDown(proto.InputMouseButtonLeft)
	p.PausaHumana(0)
	p.Mouse.MustUp(proto.InputMouseButtonLeft)
}

// RetornaElemento retorna um elemento e recebe um tempo de timeout. Retornará um erro caso acabe o tempo e o elemento não for encontrado
func (p *Pagina) RetornaElemento(HTMLElement string, tempo time.Duration) (*rod.Element, error) {
	return p.Timeout(tempo * time.Second).Element(HTMLElement)
}

// EscreverComoHumano escreve o texto de modo humanizado com delay random entre cada tecla digitada, além de clicar no campo do form como um humano faria usando a função interna ApertarElemento.
func (p *Pagina) EscreverComoHumano(HTMLElement string, conteudo string) error {
	el, err := p.RetornaElemento(HTMLElement, 5)
	if err != nil {
		return fmt.Errorf("erro ao encontrar elemento %s: %w\n", HTMLElement, err)
	}
	p.ApertarElementoDefinido(el)

	for _, char := range conteudo {
		p.PausaHumana(1)
		el.MustInput(string(char))
	}
	return nil
}

func (p *Pagina) DigitarTecladoComoHumano(HTMLElement string, conteudo string) error {
	el, err := p.RetornaElemento(HTMLElement, 5)
	if err != nil {
		return fmt.Errorf("erro ao encontrar elemento %s: %w\n", HTMLElement, err)
	}
	p.ApertarElementoDefinido(el)

	for _, char := range conteudo {
		p.PausaHumana(1)
		p.Keyboard.MustType(input.Key(char))
	}
	return nil
}

// LocalizarElemento localiza um elemento na tela. Recebe como parâmetro o tempo de timeout em segundos que poderá demorar e o elemento à se procurar.
// O correto é que se coloque "." para buscar por classe ou "#" por 'ID'.
//
// Retorna X e Y de um elemento, ou erro caso falhe em encontrar.
func (p *Pagina) LocalizarElemento(timeout time.Duration, HTMLElement string) (elX, elY float64, err error) {
	el, err := p.Timeout(timeout * time.Second).Element(HTMLElement)
	if err != nil {
		return 0, 0, err
	}
	return el.MustShape().Box().X, el.MustShape().Box().Y, nil
}

// PausaHumana cria um intervalo de pouco menos de meio segundo entre cada ação. Extremamente importante para que o navegador não identifique o 'bot'
// Recebe como argumento 0 ou 1, qualquer outro número faz com que vire 1 (hesitação humana)
// Possui 2 tipos
//
// Tipo 0: Delay do dedo descendo e subindo ao apertar um botão
//
// Tipo 1: Delay de digitação entre teclas
//
// Default: Hesitação humana: recomendado antes de apertar um botão, antes de sair da barra de pesquisa, antes de ir procurar algo
func (p *Pagina) PausaHumana(tipo uint8) {
	switch tipo {
	case 0:
		time.Sleep(time.Duration(rand.Intn(50)+50) * time.Millisecond)
	case 1:
		time.Sleep(time.Duration(rand.Intn(80)+90) * time.Millisecond)
	default:
		time.Sleep(time.Duration(rand.Intn(300)+100) * time.Millisecond)
	}
}

func (p *Pagina) DefinirPastaDownload(path string) error {
	err := proto.BrowserSetDownloadBehavior{Behavior: proto.BrowserSetDownloadBehaviorBehaviorAllow, DownloadPath: path}.Call(p)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrConfigurarPastaDownloadDefault, err)
	}
	return nil
}

// Fechar a conexão
func (p *Pagina) Fechar() {
	_ = p.Close()
}
