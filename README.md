# nfse

HTTP service that emits NFS-e (Nota Fiscal de Serviço Eletrônica) against the
São Paulo city portal at `https://nfe.prefeitura.sp.gov.br` by driving a real
Chromium browser through Login Único + the emission form. Captchas are solved
via [2Captcha](https://2captcha.com/). One JSON POST → one or more emitted
notas, each as a PDF downloaded to the local filesystem.

- HTTP entrypoint: [`cmd/main/main.go`](cmd/main/main.go)
- Workflow orchestrator: [`pkg/workflow/workflow.go`](pkg/workflow/workflow.go)
- Browser automation: [`pkg/receita/`](pkg/receita/) (Rod + stealth)
- Request/response types: [`pkg/domain/prestador.go`](pkg/domain/prestador.go)

## API

The service exposes a single endpoint:

### `POST /prestador`

Request body (JSON):

| Field | Type | Notes |
| --- | --- | --- |
| `prestador` | string | Nome do prestador. Also used as the subfolder under `NFs/` where PDFs are saved. |
| `login` | string | CPF/CNPJ used on Login Único. |
| `senha` | string | Senha do portal. |
| `notas` | `[]Nota` | One or more notas to emit. |

Each `Nota`:

| Field | Type | Notes |
| --- | --- | --- |
| `tomador` | string | Nome/razão social do tomador. |
| `cnpj` | string | CPF ou CNPJ do tomador. |
| `valor` | string | Valor do serviço (ex. `"10"`, `"1234.56"`). |
| `observacao` | string | Discriminação do serviço. |
| `data` | string | `dd/mm/yyyy`. |

Example:

```bash
curl -X POST http://localhost:8080/prestador \
  -H 'Content-Type: application/json' \
  -d '{
    "prestador": "Nome do Prestador",
    "login": "<CPF-ou-CNPJ-do-prestador>",
    "senha": "<senha-do-portal>",
    "notas": [
      {
        "tomador": "Nome do Tomador",
        "cnpj": "<CPF-ou-CNPJ-do-tomador>",
        "valor": "10",
        "observacao": "teste",
        "data": "02/03/2026"
      }
    ]
  }'
```

Responses:

- `200 OK` + body `Notas emitidas com sucesso` — every nota was emitted.
- `400 Bad Request` — JSON decode / validation error.
- `405 Method Not Allowed` — anything other than `POST`.
- `500 Internal Server Error` — workflow failure (login, captcha, portal).

Side effects on success:

- A PDF per nota is downloaded into `NFs/<prestador>/` relative to the binary.
- Append-only log file created in `logs/` relative to the binary (one per process start).

## Configuration

Environment variables:

| Var | Default | Purpose |
| --- | --- | --- |
| `PORT` | random high port | Port the HTTP server listens on. Can be `8080` or `:8080`. |
| `CHROMIUM_BIN` | `/opt/chromium/chrome-linux/chrome` | Absolute path to the Chromium binary Rod should launch. |
| `NFSE_HEADLESS` | `0` (headful) | `1`/`true` forces headless; `0`/`false` forces headful; unset = whatever the code asks for. |
| `SITE` | `https://nfe.prefeitura.sp.gov.br` | Base URL of the NFS-e portal. Used by tests to point at a mock. |
| `NFSE_SKIP_E2E` | unset | `1` skips the e2e tests in [`pkg/receita/`](pkg/receita/). |
| `NFSE_E2E_PAYLOAD_FILE` | unset | If set, `TestFullWorkflowE2E` POSTs the contents of this file instead of the committed dummy payload. |

The 2Captcha API key is currently embedded in
[`pkg/captcha/captcha.go`](pkg/captcha/captcha.go); if that key is spent or
invalid, captcha solving will fail at runtime.

## Running locally

Prereqs: **Go 1.25.4** and a Chromium binary. The simplest way to get Chromium
is to let Rod download it on the first test run and reuse the cached path.

```bash
# Build
go build -o bin/nfse ./cmd/main

# Download the pinned Chromium snapshot once (matches CI + Dockerfile).
# Use your preferred path; the default the code expects is
# /opt/chromium/chrome-linux/chrome (overridable via CHROMIUM_BIN).
CHROMIUM_REV=1321438
curl -fsSL -o /tmp/chrome.zip \
  "https://storage.googleapis.com/chromium-browser-snapshots/Linux_x64/${CHROMIUM_REV}/chrome-linux.zip"
sudo mkdir -p /opt/chromium && sudo unzip -q /tmp/chrome.zip -d /opt/chromium/
sudo chmod -R a+rx /opt/chromium/chrome-linux

# Run (port is randomized if PORT is unset; see startup log for the port).
PORT=8080 NFSE_HEADLESS=1 ./bin/nfse
```

On startup the service logs:

```
Servidor rodando na porta :8080
```

## Running in Docker

The provided `Dockerfile` is multi-stage and ships a pinned Chromium snapshot
inside the image. No external Chromium install needed on the host.

```bash
docker build -t nfse:local .
docker run --rm -e PORT=8080 -p 8080:8080 nfse:local
```

Notes:

- `NFSE_HEADLESS=1` and `CHROMIUM_BIN=/opt/chromium/chrome-linux/chrome` are
  set as image defaults.
- If you need headful (e.g. for debugging), set `NFSE_HEADLESS=0`; the
  entrypoint spins up `Xvfb` on `DISPLAY=:99` automatically.
- The runtime container runs as an unprivileged `nfse` user.

## Continuous integration

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs on every push and
PR and has three jobs:

1. **Build & vet** — `go build ./...`, `go vet ./...`, uploads a
   `nfse-linux-amd64` artifact.
2. **Unit + E2E tests** — installs Chromium runtime libs, downloads the pinned
   Chromium snapshot, runs `go test -v -timeout 5m ./...`. The e2e tests in
   `pkg/receita/` drive a real Chromium through a local `httptest` mock of the
   SP portal.
3. **Docker image build** — builds the Dockerfile and smoke-tests that the
   container starts and serves `/prestador` (expects `GET` → `405`).

## End-to-end tests

Two tests in [`pkg/receita/receita_e2e_test.go`](pkg/receita/receita_e2e_test.go):

- `TestLoginE2E` — drives Rod through the mock portal login flow and asserts
  the CPF/senha landed on the mock.
- `TestFullWorkflowE2E` — wires up the full handler stack (logger + captcha +
  ui + workflow + `/prestador` handler) in-process on a random
  `127.0.0.1` port, POSTs a JSON payload, and asserts the mock portal recorded
  one emitted nota per payload entry with the CNPJ, tomador, observação,
  valor, and data the automation typed.

Run them:

```bash
# Both tests. Needs Chromium available (see "Running locally").
CHROMIUM_BIN=/opt/chromium/chrome-linux/chrome NFSE_HEADLESS=1 \
  go test -race -v -timeout 10m ./pkg/receita/...

# Drive TestFullWorkflowE2E through a custom payload (never committed):
NFSE_E2E_PAYLOAD_FILE=/tmp/real-payload.json \
CHROMIUM_BIN=/opt/chromium/chrome-linux/chrome NFSE_HEADLESS=1 \
  go test -v -run TestFullWorkflowE2E ./pkg/receita/...

# Skip e2e entirely (unit-only):
NFSE_SKIP_E2E=1 go test ./...
# or
go test -short ./...
```

Neither test ever hits `nfe.prefeitura.sp.gov.br`; all traffic stays on
`127.0.0.1`.

## Project layout

```
cmd/main/       HTTP entrypoint
cmd/math/       Port selection (PORT env or random)
pkg/captcha/    2Captcha client
pkg/domain/     Prestador/Nota types + request validation
pkg/handlers/   /prestador HTTP handler
pkg/logger/     File-based logger (logs/nfse_<timestamp>.log)
pkg/receita/    Rod-driven SP portal automation (login + emission forms)
pkg/rod/        Rod wrapper (stealth, headless env gate, downloads)
pkg/sistema/    Filesystem helpers (NFs/<prestador>/, logs/)
pkg/ui/         CLI UI (stdout printers)
pkg/workflow/   Orchestrator called by the /prestador handler
scripts/        systemd-style install script (legacy, optional)
```

## License

See [LICENSE.md](LICENSE.md) — proprietary, all rights reserved.
