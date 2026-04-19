package receita_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"nfse/pkg/captcha"
	"nfse/pkg/handlers"
	"nfse/pkg/logger"
	"nfse/pkg/receita"
	nfserod "nfse/pkg/rod"
	"nfse/pkg/ui"
	"nfse/pkg/workflow"
)

// mockPortal stands up a tiny HTTP server that mimics just enough of the
// SP NFS-e portal for Rod to drive through it end-to-end:
//
//	1. GET /login.aspx                → page with `.oauth-button` → /login-unico
//	2. GET /login-unico               → page with input[name=Username],
//	                                    input[name=Password], and a
//	                                    button[type=submit] Entrar
//	                                    (on click, JS redirects to /logged-in)
//	3. GET /logged-in                 → landing page (no .text-danger, no captcha)
//	4. GET /contribuinte/nota.aspx    → CNPJ + data form with #ctl00_body_btAvancar
//	5. GET /contribuinte/emissao      → tomador/obs/valor form with #ctl00_body_btEmitir
//	                                    (click triggers window.confirm and on OK
//	                                    navigates to /contribuinte/emitido)
//	6. GET /contribuinte/emitido      → post-emit page with #btDownload and
//	                                    #ctl00_cphBase_btVoltar
//
// It records the login credentials and every emitted nota so tests can
// assert the automation submitted the right values.
type mockPortal struct {
	server *httptest.Server

	// mu guards the fields below, which are written from the HTTP handler
	// goroutine and read from the test goroutine.
	mu             sync.Mutex
	loggedInCpf    string
	loggedInSenha  string
	gotLoggedInHit bool
	notas          []mockNota
}

type mockNota struct {
	Cnpj       string
	Data       string
	Tomador    string
	Observacao string
	Valor      string
}

func (mp *mockPortal) snapshotLogin() (hit bool, cpf, senha string) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	return mp.gotLoggedInHit, mp.loggedInCpf, mp.loggedInSenha
}

func (mp *mockPortal) snapshotNotas() []mockNota {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	out := make([]mockNota, len(mp.notas))
	copy(out, mp.notas)
	return out
}

func newMockPortal(t *testing.T) *mockPortal {
	t.Helper()
	mp := &mockPortal{}

	mux := http.NewServeMux()

	mux.HandleFunc("/login.aspx", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html><html><head><meta charset="utf-8"><title>mock receita</title></head>
<body>
<h1>Mock Receita SP</h1>
<a class="oauth-button" href="/login-unico">Entrar com Login Único</a>
</body></html>`)
	})

	mux.HandleFunc("/login-unico", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Mirror the real Login Único form at pmspauth.prefeitura.sp.gov.br:
		// fields are addressed by `name="Username"`/`name="Password"` and the
		// submit button is a plain `button[type="submit"]`.
		fmt.Fprint(w, `<!doctype html><html><head><meta charset="utf-8"><title>mock login único</title></head>
<body>
<h1>Login Único (mock)</h1>
<form id="loginForm" onsubmit="return submitLogin();">
  <input name="Username" autocomplete="off">
  <input name="Password" type="password" autocomplete="off">
  <button type="submit" onclick="submitLogin()">Entrar</button>
</form>
<script>
function submitLogin() {
  var u = encodeURIComponent(document.getElementsByName('Username')[0].value);
  var p = encodeURIComponent(document.getElementsByName('Password')[0].value);
  window.location.href = '/logged-in?cpf=' + u + '&senha=' + p;
  return false;
}
</script>
</body></html>`)
	})

	mux.HandleFunc("/logged-in", func(w http.ResponseWriter, r *http.Request) {
		cpf := r.URL.Query().Get("cpf")
		senha := r.URL.Query().Get("senha")
		mp.mu.Lock()
		mp.gotLoggedInHit = true
		mp.loggedInCpf = cpf
		mp.loggedInSenha = senha
		mp.mu.Unlock()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html><html><head><meta charset="utf-8"><title>mock dashboard</title></head>
<body>
<h1 id="welcome">Bem-vindo</h1>
<p>Login realizado com sucesso.</p>
</body></html>`)
	})

	// CNPJ/data form. Avançar grabs both fields and navigates to the emission
	// form, carrying the values as query parameters.
	mux.HandleFunc("/contribuinte/nota.aspx", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html><html><head><meta charset="utf-8"><title>mock nota</title></head>
<body>
<h1>Emissão de NFS-e (mock)</h1>
<form id="notaForm" onsubmit="return avancar();">
  <label>CPF/CNPJ Tomador <input id="ctl00_body_tbCPFCNPJTomador" autocomplete="off"></label>
  <label>Data <input id="ctl00_body_txtEmitidoEm" autocomplete="off"></label>
  <button type="button" id="ctl00_body_btAvancar" onclick="avancar()">Avançar</button>
</form>
<script>
function avancar() {
  var cnpj = encodeURIComponent(document.getElementById('ctl00_body_tbCPFCNPJTomador').value);
  var data = encodeURIComponent(document.getElementById('ctl00_body_txtEmitidoEm').value);
  window.location.href = '/contribuinte/emissao?cnpj=' + cnpj + '&data=' + data;
  return false;
}
</script>
</body></html>`)
	})

	// Emission form. Emitir fires window.confirm first; on OK, navigates to
	// /contribuinte/emitido with every field value as a query parameter.
	mux.HandleFunc("/contribuinte/emissao", func(w http.ResponseWriter, r *http.Request) {
		cnpj := r.URL.Query().Get("cnpj")
		data := r.URL.Query().Get("data")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!doctype html><html><head><meta charset="utf-8"><title>mock emissao</title></head>
<body>
<h1>Dados da nota (mock)</h1>
<p>CNPJ: %s | Data: %s</p>
<form id="emitirForm" onsubmit="return emitir();">
  <label>Razão Social <input id="ctl00_body_tbRazaoSocial" autocomplete="off"></label>
  <label>Discriminação <input id="ctl00_body_tbDiscriminacao" autocomplete="off"></label>
  <label>Valor <input id="ctl00_body_tbValor" autocomplete="off"></label>
  <button type="button" id="ctl00_body_btEmitir" onclick="emitir()">Emitir</button>
</form>
<script>
var __cnpj = %q;
var __data = %q;
function emitir() {
  if (!confirm('Deseja emitir a NFS-e?')) { return false; }
  var tomador = encodeURIComponent(document.getElementById('ctl00_body_tbRazaoSocial').value);
  var obs     = encodeURIComponent(document.getElementById('ctl00_body_tbDiscriminacao').value);
  var valor   = encodeURIComponent(document.getElementById('ctl00_body_tbValor').value);
  window.location.href = '/contribuinte/emitido?cnpj=' + encodeURIComponent(__cnpj)
     + '&data=' + encodeURIComponent(__data)
     + '&tomador=' + tomador
     + '&obs=' + obs
     + '&valor=' + valor;
  return false;
}
</script>
</body></html>`, cnpj, data, cnpj, data)
	})

	// Post-emit page. Records the nota and serves a download button (no-op) +
	// a back button that navigates to /contribuinte/nota.aspx so the workflow
	// can loop to emit another nota.
	mux.HandleFunc("/contribuinte/emitido", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		nota := mockNota{
			Cnpj:       q.Get("cnpj"),
			Data:       q.Get("data"),
			Tomador:    q.Get("tomador"),
			Observacao: q.Get("obs"),
			Valor:      q.Get("valor"),
		}
		mp.mu.Lock()
		mp.notas = append(mp.notas, nota)
		mp.mu.Unlock()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html><html><head><meta charset="utf-8"><title>mock emitido</title></head>
<body>
<h1>NFS-e emitida (mock)</h1>
<button type="button" id="btDownload" onclick="window.__downloaded=true">Download</button>
<button type="button" id="ctl00_cphBase_btVoltar" onclick="window.location.href='/contribuinte/nota.aspx'">Voltar</button>
</body></html>`)
	})

	mp.server = httptest.NewServer(mux)
	t.Cleanup(mp.server.Close)
	return mp
}

// resolveChromium picks a Chromium binary path: $CHROMIUM_BIN, a common
// Rod cache location, or the hardcoded /opt path. If none exist, returns "".
func resolveChromium(t *testing.T) string {
	t.Helper()
	candidates := []string{os.Getenv("CHROMIUM_BIN"), "/opt/chromium/chrome-linux/chrome"}

	// Look under ~/.cache/rod/browser for a Rod-downloaded Chromium.
	if home, err := os.UserHomeDir(); err == nil {
		rodDir := home + "/.cache/rod/browser"
		if entries, err := os.ReadDir(rodDir); err == nil {
			for _, e := range entries {
				if e.IsDir() && strings.HasPrefix(e.Name(), "chromium-") {
					candidates = append(candidates, rodDir+"/"+e.Name()+"/chrome")
				}
			}
		}
	}

	for _, p := range candidates {
		if p == "" {
			continue
		}
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

func skipIfNoE2E(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("e2e test skipped in -short mode")
	}
	if os.Getenv("NFSE_SKIP_E2E") == "1" {
		t.Skip("NFSE_SKIP_E2E=1 set")
	}
	bin := resolveChromium(t)
	if bin == "" {
		t.Skip("no Chromium binary found; set CHROMIUM_BIN or install Chromium under /opt/chromium/chrome-linux/ or ~/.cache/rod/browser/")
	}
	return bin
}

// TestLoginE2E drives the real Chromium (via Rod) through a mock of the SP
// NFS-e portal login flow: acessa o site, acha e aperta o botão de login
// único, preenche CPF/senha, e confirma.
//
// The test is skipped when neither $CHROMIUM_BIN nor a default Rod-cached
// Chromium is available (e.g. on minimal dev machines). CI is responsible
// for making sure Chromium is present before running it.
func TestLoginE2E(t *testing.T) {
	bin := skipIfNoE2E(t)

	mp := newMockPortal(t)

	t.Setenv("SITE", mp.server.URL)
	t.Setenv("CHROMIUM_BIN", bin)
	t.Setenv("NFSE_HEADLESS", "1")

	pagina, err := nfserod.CriarNavegador(true)
	if err != nil {
		t.Fatalf("CriarNavegador: %v", err)
	}
	t.Cleanup(func() { _ = pagina.Navegador.Close() })

	r := receita.New(pagina, &captcha.Captcha{})

	if err := r.AcessarSiteReceita(); err != nil {
		t.Fatalf("AcessarSiteReceita: %v", err)
	}

	if err := r.EncontrarBotaoLoginUnico(); err != nil {
		t.Fatalf("EncontrarBotaoLoginUnico: %v", err)
	}

	if err := r.ApertarLoginUnico(); err != nil {
		t.Fatalf("ApertarLoginUnico: %v", err)
	}

	const cpf = "12345678900"
	const senha = "hunter2-fake"

	if err := r.ColocarDadosLogin(cpf, senha); err != nil {
		t.Fatalf("ColocarDadosLogin: %v", err)
	}

	if err := r.ApertarLogin(); err != nil {
		t.Fatalf("ApertarLogin: %v", err)
	}

	// Give the mock portal up to a few seconds to observe the /logged-in hit.
	deadline := time.Now().Add(15 * time.Second)
	var (
		gotHit   bool
		gotCpf   string
		gotSenha string
	)
	for time.Now().Before(deadline) {
		gotHit, gotCpf, gotSenha = mp.snapshotLogin()
		if gotHit {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !gotHit {
		t.Fatalf("mock portal never saw /logged-in hit after login click")
	}
	if gotCpf != cpf {
		t.Errorf("mock portal received cpf=%q, want %q", gotCpf, cpf)
	}
	if gotSenha != senha {
		t.Errorf("mock portal received senha=%q, want %q", gotSenha, senha)
	}
}

// defaultFullE2EPayload mirrors the sample payload from the user request but
// uses clearly-fake credentials so the committed test does not carry real
// login/senha. Override with $NFSE_E2E_PAYLOAD_FILE to run the test against
// an arbitrary payload locally (e.g. your real provider creds). That file is
// never read in CI.
const defaultFullE2EPayload = `{
    "prestador": "MockPrestador",
    "login": "00000000000000",
    "senha": "mock-senha",
    "notas": [
        {
            "tomador": "MockTomador",
            "cnpj": "00000000000",
            "valor": "10",
            "observacao": "teste",
            "data": "02/03/2026"
        }
    ]
}`

func loadFullE2EPayload(t *testing.T) []byte {
	t.Helper()
	if path := os.Getenv("NFSE_E2E_PAYLOAD_FILE"); path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading NFSE_E2E_PAYLOAD_FILE=%q: %v", path, err)
		}
		return b
	}
	return []byte(defaultFullE2EPayload)
}

// TestFullWorkflowE2E wires up the real HTTP handler + workflow + Rod stack
// exactly like cmd/main/main.go does, points it at the mock portal via the
// SITE env var, and POSTs a full /prestador payload. It then asserts:
//
//   - the HTTP response is 200 OK
//   - the mock portal observed the login step with the provided CPF/senha
//   - the mock portal recorded one emitted nota per entry in the payload,
//     with the CNPJ, tomador, observação, valor, and (comma-stripped) data
//     that the automation submitted
//
// This exercises the complete code path from /prestador handler down to
// Rod-driven form filling, without touching the real SP portal.
func TestFullWorkflowE2E(t *testing.T) {
	bin := skipIfNoE2E(t)

	mp := newMockPortal(t)

	t.Setenv("SITE", mp.server.URL)
	t.Setenv("CHROMIUM_BIN", bin)
	t.Setenv("NFSE_HEADLESS", "1")

	// Decode the payload so we can assert against the values the automation
	// is supposed to type into the portal.
	payload := loadFullE2EPayload(t)
	var want struct {
		Prestador string `json:"prestador"`
		Login     string `json:"login"`
		Senha     string `json:"senha"`
		Notas     []struct {
			Tomador    string `json:"tomador"`
			Cnpj       string `json:"cnpj"`
			Valor      string `json:"valor"`
			Observacao string `json:"observacao"`
			Data       string `json:"data"`
		} `json:"notas"`
	}
	if err := json.Unmarshal(payload, &want); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if len(want.Notas) == 0 {
		t.Fatalf("payload has no notas; test would not exercise the emission path")
	}

	// Construct the full handler stack, matching cmd/main/main.go.
	log, err := logger.New()
	if err != nil {
		t.Fatalf("logger.New: %v", err)
	}
	t.Cleanup(log.Fechar)
	w := workflow.New(log, captcha.New(), ui.NewUI())
	handler := handlers.NewPrestadorHandler(w)

	mux := http.NewServeMux()
	mux.HandleFunc("/prestador", handler.HandlePost)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })

	addr := ln.Addr().String()
	t.Logf("nfse server listening on http://%s/prestador", addr)
	t.Logf("mock portal listening on %s", mp.server.URL)

	// Run the POST on a goroutine with a generous timeout — driving real
	// Chromium through the full flow takes ~20-30s on CI.
	client := &http.Client{Timeout: 3 * time.Minute}
	resp, err := client.Post("http://"+addr+"/prestador", "application/json", strings.NewReader(string(payload)))
	if err != nil {
		t.Fatalf("POST /prestador: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /prestador returned %d: %s", resp.StatusCode, string(body))
	}

	// Assert login was observed with the payload's CPF/senha.
	gotHit, gotCpf, gotSenha := mp.snapshotLogin()
	if !gotHit {
		t.Fatalf("mock portal never saw /logged-in hit during the full workflow")
	}
	if gotCpf != want.Login {
		t.Errorf("mock portal received cpf=%q, want %q", gotCpf, want.Login)
	}
	if gotSenha != want.Senha {
		t.Errorf("mock portal received senha=%q, want %q", gotSenha, want.Senha)
	}

	// Assert every nota in the payload was emitted against the mock portal.
	gotNotas := mp.snapshotNotas()
	if len(gotNotas) != len(want.Notas) {
		t.Fatalf("mock portal recorded %d notas, want %d: %+v", len(gotNotas), len(want.Notas), gotNotas)
	}
	for i, wn := range want.Notas {
		gn := gotNotas[i]
		// DataParaDigitarComoHumano strips the "/" before typing, so the
		// mock sees the compact form.
		wantData := strings.ReplaceAll(wn.Data, "/", "")
		if gn.Cnpj != wn.Cnpj {
			t.Errorf("nota[%d] cnpj = %q, want %q", i, gn.Cnpj, wn.Cnpj)
		}
		if gn.Data != wantData {
			t.Errorf("nota[%d] data = %q, want %q", i, gn.Data, wantData)
		}
		if gn.Tomador != wn.Tomador {
			t.Errorf("nota[%d] tomador = %q, want %q", i, gn.Tomador, wn.Tomador)
		}
		if gn.Observacao != wn.Observacao {
			t.Errorf("nota[%d] observacao = %q, want %q", i, gn.Observacao, wn.Observacao)
		}
		if gn.Valor != wn.Valor {
			t.Errorf("nota[%d] valor = %q, want %q", i, gn.Valor, wn.Valor)
		}
	}
}
