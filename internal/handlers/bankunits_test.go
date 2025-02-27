package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pkarmon/swiftcodes/internal/handlers"
	"github.com/stretchr/testify/assert"
)

type Result struct {
	Result int `json:"result"`
}

func handleAddition(w http.ResponseWriter, r *http.Request) {
	x := mux.Vars(r)["x"]
	y := mux.Vars(r)["y"]

	a, err := strconv.Atoi(x)
	if err != nil {
		http.Error(w, "Invalid value for x", http.StatusBadRequest)
	}

	b, err := strconv.Atoi(y)
	if err != nil {
		http.Error(w, "Invalid value for y", http.StatusBadRequest)
	}

	result := a + b

	handlers.Encode(w, http.StatusOK, Result{Result: result})
}

func TestHandleAddition(t *testing.T) {
	req := httptest.NewRequest("GET", "/add/1/2", nil)
	rec := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/add/{x}/{y}", handleAddition)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	result, err := handlers.Decode[Result](io.NopCloser(rec.Body))
	assert.NoError(t, err)
	assert.Equal(t, 3, result.Result)
}
