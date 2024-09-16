package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	resultKeyData      = "data"
	resultKeyTimestamp = "timestamp"
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	resultEntriesCount := rand.Intn(10) + 1
	resultEntries := make([]map[string]interface{}, resultEntriesCount)
	for i := 0; i < resultEntriesCount; i++ {
		resultEntries[i] = generateData()
	}
	response := map[string]interface{}{
		"results": resultEntries,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("[INFO] Generated data with %v entries", resultEntriesCount)
}

func generateData() map[string]interface{} {
	return map[string]interface{}{
		resultKeyData:      fmt.Sprintf("data-%d", rand.Intn(100)),
		resultKeyTimestamp: time.Now().UnixNano(),
	}
}

func main() {
	http.HandleFunc("/handle", handler)
	fmt.Println("Server is running on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
