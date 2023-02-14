package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

type Salt struct {
	Value string `json:"salt"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/generate-salt", generateSaltHandler)
	srv := http.Server{Addr: ":8787", Handler: mux}
	log.Println("Listening on port: 8787")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func generateSaltHandler(w http.ResponseWriter, r *http.Request) {
	salt := Salt{Value: generateSalt()}
	w.Header().Set("Content-Type", "application/json")
	payload, err := json.Marshal(salt)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(payload)
}

func generateSalt() string {
	res := strings.Builder{}
	for i := 0; i < 12; i++ {
		res.WriteRune(getRandomRune())
	}
	return res.String()
}

func getRandomRune() rune {
	r := rand.Int31n(62)
	switch {
	case r >= 52:
		return '0' + r - 52
	case r >= 26:
		return 'A' + r - 26
	default:
		return 'a' + r
	}
}
