package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Response struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %+v", err)
	}

	http.HandleFunc("/balances", balancesHandler)
	http.HandleFunc("/transactions", transactionsHandler)
	http.HandleFunc("/exchange-fees", exchangeFeesHandler)
	http.HandleFunc("/exchange-pairs", exchangePairsHandler)
	http.HandleFunc("/exchange-rates", exchangeRatesHandler)
	http.HandleFunc("/withdrawal-fees", withdrawalFeesHandler)

	port := os.Getenv("PORT")
	log.Println("Starting server on :" + port + "...")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func balancesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "Missing 'userId' query parameter", http.StatusBadRequest)
		return
	}
	response := Response{
		Message: fmt.Sprintf("Balances retrieved for userId: %s", userID),
		Data:    []string{"BTC: 0.5", "ETH: 2.0", "USDT: 1500"},
	}
	respondJSON(w, response)
}

func transactionsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "Missing 'userId' query parameter", http.StatusBadRequest)
		return
	}
	response := Response{
		Message: fmt.Sprintf("Transaction history retrieved for userId: %s", userID),
		Data: []map[string]string{
			{"id": "1", "type": "deposit", "amount": "100", "currency": "USDT"},
			{"id": "2", "type": "withdrawal", "amount": "0.01", "currency": "BTC"},
		},
	}
	respondJSON(w, response)
}

func exchangeFeesHandler(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "Missing 'symbol' query parameter", http.StatusBadRequest)
		return
	}
	response := Response{
		Message: fmt.Sprintf("Exchange fees retrieved for symbol: %s", symbol),
		Data:    map[string]string{symbol: "0.1%"},
	}
	respondJSON(w, response)
}

func exchangePairsHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Message: "Available exchange pairs retrieved successfully.",
		Data:    []string{"BTC/USDT", "ETH/USDT", "BTC/ETH"},
	}
	respondJSON(w, response)
}

func exchangeRatesHandler(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "Missing 'symbol' query parameter", http.StatusBadRequest)
		return
	}
	response := Response{
		Message: fmt.Sprintf("Exchange rate retrieved for symbol: %s", symbol),
		Data:    map[string]string{symbol: "30000"},
	}
	respondJSON(w, response)
}

func withdrawalFeesHandler(w http.ResponseWriter, r *http.Request) {
	asset := r.URL.Query().Get("asset")
	if asset == "" {
		http.Error(w, "Missing 'asset' query parameter", http.StatusBadRequest)
		return
	}
	network := r.URL.Query().Get("network")
	message := fmt.Sprintf("Withdrawal fee retrieved for asset: %s", asset)
	if network != "" {
		message += fmt.Sprintf(", network: %s", network)
	}
	response := Response{
		Message: message,
		Data:    map[string]string{asset: "0.0005"},
	}
	respondJSON(w, response)
}

func respondJSON(w http.ResponseWriter, response Response) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
