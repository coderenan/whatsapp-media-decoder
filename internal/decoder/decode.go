package decoder

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"golang.org/x/crypto/hkdf"
)

func DecodeMediaHandler(w http.ResponseWriter, r *http.Request) {
	var req DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if req.MediaURL == "" || req.MediaKey == "" || req.MimeType == "" {
		http.Error(w, `{"error":"Parâmetros 'media_url', 'media_key' e 'mimetype' são obrigatórios"}`, http.StatusBadRequest)
		return
	}

	resp, err := http.Get(req.MediaURL)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Erro ao baixar mídia: %s"}`, err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	encData, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Erro ao ler mídia: %s"}`, err), http.StatusInternalServerError)
		return
	}

	mediaKey, err := base64.StdEncoding.DecodeString(req.MediaKey)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Erro ao decodificar media_key: %s"}`, err), http.StatusBadRequest)
		return
	}

	var mediaType []byte
	switch {
	case req.MimeType[:6] == "image/":
		mediaType = []byte("WhatsApp Image Keys")
	case req.MimeType[:6] == "audio/":
		mediaType = []byte("WhatsApp Audio Keys")
	case req.MimeType[:6] == "video/":
		mediaType = []byte("WhatsApp Video Keys")
	case req.MimeType == "application/pdf" || req.MimeType[:11] == "application":
		mediaType = []byte("WhatsApp Document Keys")
	default:
		http.Error(w, fmt.Sprintf(`{"error":"Tipo de mídia não suportado: %s"}`, req.MimeType), http.StatusBadRequest)
		return
	}
	

	hkdfReader := hkdf.New(sha256.New, mediaKey, nil, mediaType)
	expandedKey := make([]byte, 112)
	if _, err := io.ReadFull(hkdfReader, expandedKey); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Erro ao derivar chave: %s"}`, err), http.StatusInternalServerError)
		return
	}

	iv := expandedKey[0:16]
	cipherKey := expandedKey[16:48]
	ciphertext := encData[:len(encData)-10]

	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Erro ao criar cipher: %s"}`, err), http.StatusInternalServerError)
		return
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	padLen := int(plaintext[len(plaintext)-1])
	if padLen > 16 {
		http.Error(w, `{"error":"Padding inválido"}`, http.StatusInternalServerError)
		return
	}
	plaintext = plaintext[:len(plaintext)-padLen]

	base64Media := base64.StdEncoding.EncodeToString(plaintext)

	respJSON := DecodeResponse{
		Success: true,
		Base64:  base64Media,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respJSON)
}
