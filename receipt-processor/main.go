package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"github.com/google/uuid"
)

type Receipt struct {
	ID           string `json:"id"`
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

var receipts = make(map[string]Receipt)

func calculatePoints(receipt Receipt) int {
	points := 0
	points += len(receipt.Retailer)

	total, _ := strconv.ParseFloat(receipt.Total, 64)
	if strings.HasSuffix(receipt.Total, ".00") {
		points += 50
	}

	if math.Mod(total, 0.25) == 0 {
		points += 25
	}

	points += len(receipt.Items) / 2 * 5

	for _, item := range receipt.Items {
		description := strings.TrimSpace(item.ShortDescription)
		if len(description)%3 == 0 {
			price, _ := strconv.ParseFloat(item.Price, 64)
			points += int(math.Ceil(price * 0.2))
		}
	}

	if total > 10.00 {
		points += 5
	}

	purchaseDate := strings.Split(receipt.PurchaseDate, "-")
	day, _ := strconv.Atoi(purchaseDate[2])
	if day%2 != 0 {
		points += 6
	}

	purchaseTime := strings.Split(receipt.PurchaseTime, ":")
	hour, _ := strconv.Atoi(purchaseTime[0])
	if hour >= 14 && hour < 16 {
		points += 10
	}

	return points
}

func processReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt Receipt
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&receipt)
	if err != nil {
		http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	receipt.ID = uuid.New().String()
	receipts[receipt.ID] = receipt
	fmt.Println("Stored receipt with ID:", receipt.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": receipt.ID})
}

func getPoints(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/receipts/")
	id = strings.TrimSuffix(id, "/points")

	fmt.Println("Looking up receipt with ID:", id)

	receipt, exists := receipts[id]
	if !exists {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	points := calculatePoints(receipt)
	fmt.Printf("Receipt ID: %s, Points: %d\n", id, points)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"points": points})
}

func main() {
	http.HandleFunc("/receipts/process", processReceipt)
	http.HandleFunc("/receipts/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/points") {
			getPoints(w, r)
			return
		}
		http.NotFound(w, r)
	})

	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
