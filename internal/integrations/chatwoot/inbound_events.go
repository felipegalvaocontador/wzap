package chatwoot

import (
	"context"
	"fmt"

	"wzap/internal/logger"
)

func (s *Service) handleReceipt(ctx context.Context, cfg *ChatwootConfig, payload []byte) {
	data, err := parseReceiptPayload(payload)
	if err != nil {
		return
	}

	if len(data.MessageIDs) == 0 {
		return
	}

	client := s.clientFn(cfg)
	for _, msgID := range data.MessageIDs {
		if msgID == "" {
			continue
		}

		msg, err := s.msgRepo.FindByID(ctx, cfg.SessionID, msgID)
		if err != nil {
			continue
		}

		if msg.CWConversationID == nil || *msg.CWConversationID == 0 {
			continue
		}

		if msg.CWSourceID != nil {
			_ = client.UpdateLastSeen(ctx, fmt.Sprintf("%d", cfg.InboxID), *msg.CWSourceID, *msg.CWConversationID)
		}
	}
}

func (s *Service) handleDelete(ctx context.Context, cfg *ChatwootConfig, payload []byte) {
	data, err := parseDeletePayload(payload)
	if err != nil {
		return
	}

	msgID := data.MessageID
	if msgID == "" {
		return
	}

	msg, err := s.msgRepo.FindByID(ctx, cfg.SessionID, msgID)
	if err != nil {
		return
	}

	if msg.CWMessageID == nil || msg.CWConversationID == nil {
		return
	}

	client := s.clientFn(cfg)
	if err := client.DeleteMessage(ctx, *msg.CWConversationID, *msg.CWMessageID); err != nil {
		logger.Warn().Err(err).Msg("Failed to delete Chatwoot message")
	}
}

func (s *Service) handleConnected(ctx context.Context, cfg *ChatwootConfig, payload []byte) {
	convID, err := s.findOrCreateBotConversation(ctx, cfg)
	if err != nil {
		logger.Warn().Err(err).Str("session", cfg.SessionID).Msg("Failed to find or create bot conversation for connected event")
		return
	}

	client := s.clientFn(cfg)
	_, _ = client.CreateMessage(ctx, convID, MessageReq{
		Content:     "✅ Session connected",
		MessageType: "outgoing",
	})
}

func (s *Service) handleDisconnected(ctx context.Context, cfg *ChatwootConfig, payload []byte) {
	convID, err := s.findOrCreateBotConversation(ctx, cfg)
	if err != nil {
		logger.Warn().Err(err).Str("session", cfg.SessionID).Msg("Failed to find or create bot conversation for disconnected event")
		return
	}

	client := s.clientFn(cfg)
	_, _ = client.CreateMessage(ctx, convID, MessageReq{
		Content:     "⚠️ Session disconnected",
		MessageType: "outgoing",
	})
}

func (s *Service) handleQR(ctx context.Context, cfg *ChatwootConfig, payload []byte) {
	var data struct {
		Codes []string `json:"Codes"`
	}
	if err := parseEnvelopeData(payload, &data); err != nil {
		return
	}

	if len(data.Codes) == 0 {
		return
	}

	qrContent := data.Codes[0]

	convID, err := s.findOrCreateBotConversation(ctx, cfg)
	if err != nil {
		logger.Warn().Err(err).Str("session", cfg.SessionID).Msg("Failed to find or create bot conversation for QR event")
		return
	}

	client := s.clientFn(cfg)
	qrPNG, err := generateQRCodePNG(qrContent)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to generate QR code")
		return
	}

	_, _ = client.CreateMessageWithAttachment(ctx, convID, "Scan QR code to connect", "qrcode.png", qrPNG, "image/png", "outgoing", "")
}

func (s *Service) handleContact(ctx context.Context, cfg *ChatwootConfig, payload []byte) {
	var data struct {
		JID       string `json:"JID"`
		FirstName string `json:"FirstName"`
		FullName  string `json:"FullName"`
	}
	if err := parseEnvelopeData(payload, &data); err != nil {
		return
	}

	if data.JID == "" {
		return
	}

	phone := extractPhone(data.JID)
	name := data.FullName
	if name == "" {
		name = data.FirstName
	}
	if name == "" {
		return
	}

	client := s.clientFn(cfg)
	contacts, _ := client.FilterContacts(ctx, phone)
	if len(contacts) == 0 {
		return
	}

	_ = client.UpdateContact(ctx, contacts[0].ID, UpdateContactReq{Name: name})
}

func (s *Service) handlePushName(ctx context.Context, cfg *ChatwootConfig, payload []byte) {
	var data struct {
		JID      string `json:"JID"`
		PushName string `json:"PushName"`
	}
	if err := parseEnvelopeData(payload, &data); err != nil {
		return
	}

	if data.JID == "" || data.PushName == "" {
		return
	}

	phone := extractPhone(data.JID)

	client := s.clientFn(cfg)
	contacts, _ := client.FilterContacts(ctx, phone)
	if len(contacts) == 0 {
		return
	}

	_ = client.UpdateContact(ctx, contacts[0].ID, UpdateContactReq{Name: data.PushName})
}

func (s *Service) handlePicture(ctx context.Context, cfg *ChatwootConfig, payload []byte) {
	var data struct {
		JID       string `json:"JID"`
		PictureID string `json:"ID"`
		URL       string `json:"URL"`
		IsGroup   bool   `json:"IsGroup"`
	}
	if err := parseEnvelopeData(payload, &data); err != nil {
		return
	}

	if data.JID == "" || data.URL == "" {
		return
	}

	phone := extractPhone(data.JID)

	client := s.clientFn(cfg)
	contacts, _ := client.FilterContacts(ctx, phone)
	if len(contacts) == 0 {
		return
	}

	_ = client.UpdateContact(ctx, contacts[0].ID, UpdateContactReq{
		AdditionalAttributes: map[string]any{"avatar_url": data.URL},
	})
}
