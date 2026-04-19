package receita_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"nfse/pkg/captcha"
	nfserod "nfse/pkg/rod"
	"nfse/pkg/receita"
)

// mockPortal stands up a tiny HTTP server that mimics just enough of the
// SP NFS-e portal login flow for Rod to drive through it end-to-end:
//
//  1. GET /login.aspx               → page with `.oauth-button` → /login-unico
//  2. GET /login-unico              → page with #cpfCnpj, #password, .btn-entrar
//                                     (on click, JS redirects to /logged-in)
//  3. GET /logged-in                → landing page (no .text-danger, no captcha)
//
// It records the login credentials that were submitted via the #cpfCnpj and
// #password fields so the test can assert the automation typed them correctly.
type mockPortal struct {
	server         *httptest.Server
	loggedInCpf    string
	loggedInSenha  string
	gotLoggedInHit bool
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
		fmt.Fprint(w, `<!doctype html><html><head><meta charset="utf-8"><title>mock login único</title></head>
<body>
<h1>Login Único (mock)</h1>
<form id="loginForm" onsubmit="return submitLogin();">
  <input id="cpfCnpj" name="cpfCnpj" autocomplete="off">
  <input id="password" name="password" type="password" autocomplete="off">
  <button type="button" class="btn-entrar" onclick="submitLogin()">Entrar</button>
</form>
<script>
function submitLogin() {
  var u = encodeURIComponent(document.getElementById('cpfCnpj').value);
  var p = encodeURIComponent(document.getElementById('password').value);
  window.location.href = '/logged-in?cpf=' + u + '&senha=' + p;
  return false;
}
</script>
</body></html>`)
	})

	mux.HandleFunc("/logged-in", func(w http.ResponseWriter, r *http.Request) {
		mp.gotLoggedInHit = true
		mp.loggedInCpf = r.URL.Query().Get("cpf")
		mp.loggedInSenha = r.URL.Query().Get("senha")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html><html><head><meta charset="utf-8"><title>mock dashboard</title></head>
<body>
<h1 id="welcome">Bem-vindo</h1>
<p>Login realizado com sucesso.</p>
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

// TestLoginE2E drives the real Chromium (via Rod) through a full mock of
// the SP NFS-e portal login flow: acessa o site, acha e aperta o botão
// de login único, preenche CPF/senha, e confirma.
//
// The test is skipped when neither $CHROMIUM_BIN nor a default Rod-cached
// Chromium is available (e.g. on minimal dev machines). CI is responsible
// for making sure Chromium is present before running it.
func TestLoginE2E(t *testing.T) {
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
	for time.Now().Before(deadline) && !mp.gotLoggedInHit {
		time.Sleep(100 * time.Millisecond)
	}
	if !mp.gotLoggedInHit {
		t.Fatalf("mock portal never saw /logged-in hit after login click")
	}
	if mp.loggedInCpf != cpf {
		t.Errorf("mock portal received cpf=%q, want %q", mp.loggedInCpf, cpf)
	}
	if mp.loggedInSenha != senha {
		t.Errorf("mock portal received senha=%q, want %q", mp.loggedInSenha, senha)
	}
}
