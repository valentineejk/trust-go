package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/valentineejk/trust-go/internal/transactions"
)

type Handler struct {
	Parser transactions.Parser
}

func NewHandler(parser transactions.Parser) *Handler {
	return &Handler{
		Parser: parser,
	}
}

func (p *Handler) GetCurrentBlockHandler(w http.ResponseWriter, r *http.Request) {
	block := p.Parser.GetCurrentBlock()
	response := map[string]int{"current_block": block}
	json.NewEncoder(w).Encode(response)
}

func (p *Handler) SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if payload.Address == "" {
		http.Error(w, "Address is required", http.StatusBadRequest)
		return
	}

	subscribed := p.Parser.Subscribe(payload.Address)
	response := map[string]bool{"subscribed": subscribed}
	json.NewEncoder(w).Encode(response)
}

func (p *Handler) GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Address is required", http.StatusBadRequest)
		return
	}
	address = strings.ToLower(address)

	transactions := p.Parser.GetTransactions(address)
	json.NewEncoder(w).Encode(transactions)
}
