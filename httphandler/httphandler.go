package httphandler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Handler for http requests
type Handler struct {
	mux    *mux.Router
	logger *zap.SugaredLogger
}

// New http handler
func New(s *mux.Router, logger *zap.SugaredLogger) *Handler {
	h := Handler{s, logger}
	h.registerRoutes()

	return &h
}

type Bean struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Roaster string `json:"roaster"`
}

type Beans []Bean
type BeanResponse struct {
	Bean Bean `json:"bean"`
}
type BeansResponse struct {
	Beans Beans `json:"beans"`
}

// RegisterRoutes for all http endpoints
func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/beans", h.beans)
	h.mux.HandleFunc("/beans/{id}", h.bean)
}

func (h *Handler) beans(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")

	b := Beans{
		Bean{ID: 0, Name: "French Roast", Roaster: "Stumptown"},
	}
	json.NewEncoder(w).Encode(BeansResponse{b})
}

func (h *Handler) bean(w http.ResponseWriter, r *http.Request) {
	beanID, _ := strconv.Atoi(strings.Trim(r.URL.String(), "/beans/"))

	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")

	b := Bean{ID: beanID, Name: "French Roast", Roaster: "Stumptown"}

	json.NewEncoder(w).Encode(BeanResponse{b})
}
