package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"whatsapp-media-decoder-go/internal/decoder"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: arquivo .env não carregado")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	r := mux.NewRouter()
	r.HandleFunc("/decode", decoder.AuthMiddleware(decoder.DecodeMediaHandler)).Methods("POST")

	fmt.Printf("Servidor rodando em http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}