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
	"os"
	"os/exec"
	"strings"
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
	case strings.HasPrefix(req.MimeType, "image/"):
		mediaType = []byte("WhatsApp Image Keys")
	case strings.HasPrefix(req.MimeType, "audio/"):
		mediaType = []byte("WhatsApp Audio Keys")
	case strings.HasPrefix(req.MimeType, "video/"):
		mediaType = []byte("WhatsApp Video Keys")
	case req.MimeType == "application/pdf" || strings.HasPrefix(req.MimeType, "application/"):
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

	// Verifica se plaintext não está vazio antes de acessar o último byte
	if len(plaintext) == 0 {
		http.Error(w, `{"error":"Plaintext vazio após descriptografia"}`, http.StatusInternalServerError)
		return
	}
	padLen := int(plaintext[len(plaintext)-1])
	if padLen <= 0 || padLen > 16 || padLen > len(plaintext) {
		http.Error(w, `{"error":"Padding inválido"}`, http.StatusInternalServerError)
		return
	}
	plaintext = plaintext[:len(plaintext)-padLen]

	convertAudio := os.Getenv("CONVERT_AUDIO_TO_MP3") == "true"

	if strings.HasPrefix(req.MimeType, "audio/") && convertAudio {
		tmpInput, err := os.CreateTemp("", "audio-*.ogg")
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Erro ao criar arquivo temporário: %s"}`, err), http.StatusInternalServerError)
			return
		}
		defer os.Remove(tmpInput.Name())

		if _, err := tmpInput.Write(plaintext); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Erro ao escrever arquivo .ogg: %s"}`, err), http.StatusInternalServerError)
			return
		}
		tmpInput.Close()

		tmpOutput, err := os.CreateTemp("", "audio-*.mp3")
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Erro ao criar arquivo de saída .mp3: %s"}`, err), http.StatusInternalServerError)
			return
		}
		tmpOutput.Close()
		defer os.Remove(tmpOutput.Name())

		cmd := exec.Command("ffmpeg", "-y", "-i", tmpInput.Name(), "-acodec", "libmp3lame", tmpOutput.Name())
		if err := cmd.Run(); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Erro na conversão ffmpeg: %s"}`, err), http.StatusInternalServerError)
			return
		}

		mp3Data, err := os.ReadFile(tmpOutput.Name())
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Erro ao ler arquivo MP3: %s"}`, err), http.StatusInternalServerError)
			return
		}

		base64Media := base64.StdEncoding.EncodeToString(mp3Data)
		respJSON := DecodeResponse{
			Success: true,
			Base64:  base64Media,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(respJSON)
		return
	}

	base64Media := base64.StdEncoding.EncodeToString(plaintext)
	respJSON := DecodeResponse{
		Success: true,
		Base64:  base64Media,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respJSON)
}
