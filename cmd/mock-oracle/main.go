package main

import (
	"encoding/json"
	"hash/fnv"
	"log"
	"net/http"
	"strings"
)

func main() {
	http.HandleFunc("/prices/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/prices/")
		if id == "" {
			http.Error(w, "item ID is required", http.StatusBadRequest)
			return
		}
		h := fnv.New64a()
		_, _ = h.Write([]byte(id))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]uint64{"base_price": 100 + h.Sum64()%9901})
	})
	log.Print("mock price oracle listening on :8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}
