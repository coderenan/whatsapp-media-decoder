package decoder

type DecodeRequest struct {
	MediaURL string `json:"media_url"`
	MediaKey string `json:"media_key"`
	MimeType string `json:"mimetype"`
}

type DecodeResponse struct {
	Success bool   `json:"success"`
	Base64  string `json:"base64,omitempty"`
	Error   string `json:"error,omitempty"`
}
