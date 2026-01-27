package handlers

import (
	"encoding/json"
	"net/http"
)

func clienteHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var cliente Cliente

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&cliente)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Aqui o JSON já foi convertido para struct com sucesso
	// Você pode debugar, validar, salvar, etc

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Requisição recebida, mas retornando 400 conforme especificado"))
}
