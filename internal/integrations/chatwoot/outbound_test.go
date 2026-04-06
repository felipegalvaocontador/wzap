package chatwoot

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendAttachmentToWhatsApp_ChunkedBodyTooLarge(t *testing.T) {
	oldMax := maxMediaBytes
	maxMediaBytes = 1024
	defer func() {
		maxMediaBytes = oldMax
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flusher, _ := w.(http.Flusher)
		chunk := bytes.Repeat([]byte("a"), 256)
		for i := 0; i < 10; i++ {
			_, _ = w.Write(chunk)
			if flusher != nil {
				flusher.Flush()
			}
		}
	}))
	defer server.Close()

	svc := &Service{
		httpClient: server.Client(),
	}
	cfg := &ChatwootConfig{
		SessionID:           "sess",
		TimeoutMediaSeconds: 2,
		TimeoutLargeSeconds: 2,
	}

	_, err := svc.sendAttachmentToWhatsApp(context.Background(), cfg, "5511999999999@s.whatsapp.net", server.URL+"/file.bin", "", "file", nil)
	if err == nil {
		t.Fatal("expected attachment too large error")
	}
	if !strings.Contains(err.Error(), "attachment too large") {
		t.Fatalf("expected attachment too large error, got: %v", err)
	}
}
