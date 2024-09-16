package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("[INFO] Request body: %s", body)

	// returning a single empty object
	response := map[string]interface{}{
		"results": []map[string]string{{}},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/handle", handler)
	fmt.Println("Server is running on port 8082...")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
