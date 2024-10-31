package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Text string `json:"text"`
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	log.Println("[info] Request received")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := Response{
		Text: "Заказ в обработке..",
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Println("[info] Request processed successfully")
}

func main() {
	log.Println("[info] Cart service starting..")
	http.HandleFunc("/status", getStatus)
	log.Println("[info] Cart service started on port 8080")
	http.ListenAndServe(":8080", nil)
}
