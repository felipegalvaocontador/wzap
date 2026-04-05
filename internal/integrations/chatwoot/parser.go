package chatwoot

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"wzap/internal/model"
)

type waMessageInfo struct {
	Chat     string `json:"Chat"`
	Sender   string `json:"Sender"`
	IsFromMe bool   `json:"IsFromMe"`
	IsGroup  bool   `json:"IsGroup"`
	ID       string `json:"ID"`
	PushName string `json:"PushName"`
}

type waMessagePayload struct {
	Info    waMessageInfo          `json:"Info"`
	Message map[string]interface{} `json:"Message"`
}

type waReceiptPayload struct {
	Type       string   `json:"Type"`
	MessageIDs []string `json:"MessageIDs"`
	Chat       string   `json:"Chat"`
	Sender     string   `json:"Sender"`
	Timestamp  int64    `json:"Timestamp"`
}

type waDeletePayload struct {
	Chat      string `json:"Chat"`
	Sender    string `json:"Sender"`
	MessageID string `json:"MessageID"`
	Timestamp int64  `json:"Timestamp"`
}

func parseEnvelopeData(payload []byte, target interface{}) error {
	envelope, err := model.ParseEventEnvelope(payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal event envelope: %w", err)
	}
	if err := json.Unmarshal(envelope.Data, target); err != nil {
		return fmt.Errorf("failed to unmarshal envelope data: %w", err)
	}
	return nil
}

func parseMessagePayload(payload []byte) (*waMessagePayload, error) {
	var data waMessagePayload
	if err := parseEnvelopeData(payload, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func parseReceiptPayload(payload []byte) (*waReceiptPayload, error) {
	var data waReceiptPayload
	if err := parseEnvelopeData(payload, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func parseDeletePayload(payload []byte) (*waDeletePayload, error) {
	var data waDeletePayload
	if err := parseEnvelopeData(payload, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

type mediaInfo struct {
	DirectPath    string
	MediaKey      []byte
	FileSHA256    []byte
	FileEncSHA256 []byte
	FileLength    int
	MimeType      string
	MediaType     string
	FileName      string
}

func extractMediaInfo(msg map[string]interface{}) *mediaInfo {
	if msg == nil {
		return nil
	}

	mediaTypes := map[string]string{
		"imageMessage":    "image",
		"videoMessage":    "video",
		"audioMessage":    "audio",
		"documentMessage": "document",
		"stickerMessage":  "sticker",
	}

	for key, mt := range mediaTypes {
		sub := getMapField(msg, key)
		if sub == nil {
			continue
		}

		directPath := getStringField(sub, "directPath")
		if directPath == "" {
			continue
		}

		mediaKeyB64 := getStringField(sub, "mediaKey")
		fileSHA256B64 := getStringField(sub, "fileSHA256")
		fileEncSHA256B64 := getStringField(sub, "fileEncSHA256")

		mediaKey, _ := base64.StdEncoding.DecodeString(mediaKeyB64)
		fileSHA256, _ := base64.StdEncoding.DecodeString(fileSHA256B64)
		fileEncSHA256, _ := base64.StdEncoding.DecodeString(fileEncSHA256B64)

		mimeType := getStringField(sub, "mimetype")
		fileLength := int(getFloatField(sub, "fileLength"))
		fileName := getStringField(sub, "fileName")

		return &mediaInfo{
			DirectPath:    directPath,
			MediaKey:      mediaKey,
			FileSHA256:    fileSHA256,
			FileEncSHA256: fileEncSHA256,
			FileLength:    fileLength,
			MimeType:      mimeType,
			MediaType:     mt,
			FileName:      fileName,
		}
	}

	if dwc := getMapField(msg, "documentWithCaptionMessage"); dwc != nil {
		if innerMsg := getMapField(dwc, "message"); innerMsg != nil {
			return extractMediaInfo(innerMsg)
		}
	}

	return nil
}

func extractTextFromMessage(msg map[string]interface{}) string {
	if msg == nil {
		return ""
	}

	if conversation := getStringField(msg, "conversation"); conversation != "" {
		return conversation
	}

	if extText := getMapField(msg, "extendedTextMessage"); extText != nil {
		if text := getStringField(extText, "text"); text != "" {
			return text
		}
	}

	if imgMsg := getMapField(msg, "imageMessage"); imgMsg != nil {
		return getStringField(imgMsg, "caption")
	}

	if vidMsg := getMapField(msg, "videoMessage"); vidMsg != nil {
		return getStringField(vidMsg, "caption")
	}

	if docMsg := getMapField(msg, "documentMessage"); docMsg != nil {
		caption := getStringField(docMsg, "caption")
		filename := getStringField(docMsg, "fileName")
		if caption != "" {
			return caption
		}
		return filename
	}

	if dwc := getMapField(msg, "documentWithCaptionMessage"); dwc != nil {
		if innerMsg := getMapField(dwc, "message"); innerMsg != nil {
			if docMsg := getMapField(innerMsg, "documentMessage"); docMsg != nil {
				caption := getStringField(docMsg, "caption")
				if caption != "" {
					return caption
				}
				return getStringField(docMsg, "fileName")
			}
		}
	}

	if locMsg := getMapField(msg, "locationMessage"); locMsg != nil {
		lat := getFloatField(locMsg, "degreesLatitude")
		lng := getFloatField(locMsg, "degreesLongitude")
		name := getStringField(locMsg, "name")
		if name != "" {
			return fmt.Sprintf("📍 %s\nhttps://www.google.com/maps?q=%f,%f", name, lat, lng)
		}
		return fmt.Sprintf("📍 Location\nhttps://www.google.com/maps?q=%f,%f", lat, lng)
	}

	if contactMsg := getMapField(msg, "contactMessage"); contactMsg != nil {
		if vcard := getStringField(contactMsg, "vcard"); vcard != "" {
			return vcard
		}
		return getStringField(contactMsg, "displayName")
	}

	return ""
}

func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFloatField(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

func getMapField(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if m2, ok := v.(map[string]interface{}); ok {
			return m2
		}
	}
	return nil
}

func extractStanzaID(msg map[string]interface{}) string {
	if msg == nil {
		return ""
	}

	if ci := getMapField(msg, "contextInfo"); ci != nil {
		if id := getStringField(ci, "stanzaId"); id != "" {
			return id
		}
	}

	msgTypes := []string{
		"extendedTextMessage",
		"imageMessage",
		"videoMessage",
		"audioMessage",
		"documentMessage",
		"stickerMessage",
	}

	for _, key := range msgTypes {
		sub := getMapField(msg, key)
		if sub == nil {
			continue
		}
		ci := getMapField(sub, "contextInfo")
		if ci == nil {
			continue
		}
		if id := getStringField(ci, "stanzaId"); id != "" {
			return id
		}
	}

	return ""
}
