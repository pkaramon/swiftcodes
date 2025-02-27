package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func Encode[T any](w http.ResponseWriter, status int, data T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func Decode[T any](r io.ReadCloser) (T, error) {
	defer r.Close()
	var data T
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return data, fmt.Errorf("decode json: %w", err)
	}
	return data, nil
}

func SendErrorMsg(w http.ResponseWriter, status int, msg string) {
	type err struct {
		Error string `json:"error"`
	}
	e := err{Error: msg}
	if err := Encode(w, status, e); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func SendSuccessMsg(w http.ResponseWriter, status int, msg string) {
	type message struct {
		Message string `json:"message"`
	}

	m := message{Message: msg}
	if err := Encode(w, status, m); err != nil {
		SendServerError(w)
	}
}

func SendServerError(w http.ResponseWriter) {
	SendErrorMsg(w, http.StatusInternalServerError, "server error")
}
