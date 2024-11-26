package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type CustomerRequest struct {
	RequestText string `json:"requestText"`
}

type GatewayResponse struct {
	ResponseText string `json:"responseText"`
}

func main() {
	http.HandleFunc("/gateway", gatewayHandler)

	log.Println("Gateway service running on :8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func gatewayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	var customerRequest CustomerRequest
	err := json.NewDecoder(r.Body).Decode(&customerRequest)
	if err != nil || customerRequest.RequestText == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var response GatewayResponse
	switch customerRequest.RequestText {
	case "what is my balance of BTC?":
		response = GatewayResponse{ResponseText: "Your balance of BTC is 0.5 BTC."}
	case "what is the withdrawal fee for ETH?":
		response = GatewayResponse{ResponseText: "The withdrawal fee for ETH is 0.005 ETH."}
	case "what are the exchange rates for BTC/USDT?":
		response = GatewayResponse{ResponseText: "The exchange rate for BTC/USDT is 50000 USDT."}
	default:
		response = GatewayResponse{ResponseText: "I'm sorry, I cannot process that request."}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
