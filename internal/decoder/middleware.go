package decoder

import (
	"net/http"
	"os"
	"strings"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		secret := os.Getenv("AUTH_SECRET")

		if secret == "" {
			http.Error(w, `{"error":"Configuração inválida: AUTH_SECRET não definido"}`, http.StatusInternalServerError)
			return
		}

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"Authorization inválido ou ausente"}`, http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		if token != secret {
			http.Error(w, `{"error":"Token não autorizado"}`, http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
