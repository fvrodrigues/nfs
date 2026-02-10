package handlers

import (
	"encoding/json"
	"net/http"
	"nfse/pkg/domain"
	"nfse/pkg/random"
	"nfse/pkg/workflow"
)

type PrestadorHandler struct {
	workflow *workflow.Workflow
}

func NewPrestadorHandler(w *workflow.Workflow) *PrestadorHandler {
	return &PrestadorHandler{workflow: w}
}

func (h *PrestadorHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var prestador domain.Prestador

	if err := json.NewDecoder(r.Body).Decode(&prestador); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	ReqID := random.NewReqID()
	err := h.workflow.Executar(prestador, ReqID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Notas emitidas com sucesso"))
}
