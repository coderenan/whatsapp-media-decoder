# 📦 WhatsApp Media Decoder (Go)

Este projeto é uma API escrita em Go que permite **descriptografar mídias criptografadas do WhatsApp** (imagens, áudios, vídeos e documentos) a partir de uma `media_key` e uma `media_url`.

É útil para quem utiliza bibliotecas como [Baileys](https://github.com/adiwajshing/Baileys), [open-wa](https://openwa.dev), ou outras APIs não oficiais que fornecem acesso a mídias criptografadas.

---

## Funcionalidades

- Faz download de mídias criptografadas a partir de uma URL
- Usa a `media_key` para derivar as chaves de decodificação (HKDF + AES)
- Suporta os principais tipos de mídia: imagens, áudios, vídeos e documentos
- Retorna o arquivo descriptografado em formato **base64**
- As requisições **exigem um header Authorization** com uma chave configurável via variável de ambiente

---

## Como funciona

O WhatsApp criptografa todas as mídias usando chaves derivadas via HKDF. Para descriptografar:

1. Você precisa da `media_key` fornecida por uma API de WhatsApp.
2. A mídia é baixada usando a `media_url` (um link temporário para o arquivo criptografado).
3. O serviço deriva a chave de descriptografia, remove o MAC e padding, e retorna o conteúdo em base64.
4. Para segurança, é necessário enviar um header Authorization com o token correto configurado.

---

## Como usar

### Usando Docker Compose (Recomendado)

1. **Clone o projeto:**
   ```sh
   git clone https://github.com/coderenan/whatsapp-media-decoder.git
   cd whatsapp-media-decoder
   ```

2. **Configure as variáveis de ambiente:**
   - Edite o arquivo `.env` ou ajuste as variáveis diretamente no [`docker-compose.yml`](docker-compose.yml).
   - Exemplo de `.env`:
     ```
     AUTH_SECRET=SecretToken
     PORT=8091
     CONVERT_AUDIO_TO_MP3=true
     ```

3. **Suba o serviço:**
   ```sh
   docker compose up --build
   ```
   O serviço ficará disponível em `http://localhost:8091`.

---

### Usando Go localmente (opcional)

1. **Pré-requisitos:** Go 1.22 ou superior

2. **Configure o `.env`:**
   ```
   AUTH_SECRET=SecretToken
   PORT=8091
   CONVERT_AUDIO_TO_MP3=true
   ```

3. **Execute:**
   ```sh
   go mod tidy
   go run ./cmd/server
   ```

---

## Testando

### Endpoint: `/decode`

- **Método:** `POST`
- **Conteúdo:** `application/json`
- **Autorização:** Header `Authorization: Bearer <seu token do env>`

#### Payload (JSON):

```json
{
  "media_url": "https://mmg.whatsapp.net/d/...",
  "media_key": "media_key",
  "mimetype": "mimetype"
}
```

| Tipo      | MIME                      | Contexto HKDF            |
| --------- | ------------------------- | ------------------------ |
| Imagem    | `image/jpeg`, `image/png` | `WhatsApp Image Keys`    |
| Áudio     | `audio/ogg`, `audio/mp4`  | `WhatsApp Audio Keys`    |
| Vídeo     | `video/mp4`               | `WhatsApp Video Keys`    |
| Documento | `application/pdf`, `...`  | `WhatsApp Document Keys` |

#### Resposta esperada:
```json
{
  "success": true,
  "base64": "arquivo_em_base64"
}
```

### Testando via cURL

```sh
curl -X POST http://localhost:8091/decode \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <SecretToken>" \
  -d '{
    "media_url": "https://mmg.whatsapp.net/...",
    "media_key": "media_key",
    "mimetype": "mimetype"
  }'
```

---

## Observações

- Para converter áudios OGG para MP3, o serviço utiliza o `ffmpeg`, já instalado automaticamente no container Docker.
- O serviço só aceitará requisições com o token correto definido em `AUTH_SECRET`.

---

## Aviso legal

Este projeto não viola os termos de uso do WhatsApp, desde que:

- A media_key e media_url sejam obtidas de forma legítima via bibliotecas autorizadas ou que o usuário controle.
- Você não compartilhe arquivos reais, chaves, nem conteúdo sensível.
- Não utilize este projeto para interceptar, espionar ou invadir a privacidade de terceiros.

---

### Licença

Este projeto está licenciado sob a licença MIT.

### Contribuições

Contribuições são bem-vindas!  
Sinta-se à vontade para abrir issues, forks e pull requests.