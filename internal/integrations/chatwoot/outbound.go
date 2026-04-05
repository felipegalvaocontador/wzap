package chatwoot

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strings"

	"wzap/internal/dto"
	"wzap/internal/logger"
)

func (s *Service) HandleIncomingWebhook(ctx context.Context, sessionID string, body dto.ChatwootWebhookPayload) error {
	cfg, err := s.repo.FindBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to load chatwoot config: %w", err)
	}
	if !cfg.Enabled {
		return nil
	}

	if body.Private {
		return nil
	}

	msg := body.GetMessage()

	if msg != nil && msg.IsOutgoing() {
		return s.handleOutgoingMessage(ctx, cfg, body)
	}

	eventType := body.EventType
	if eventType == "" {
		eventType = body.Event
	}

	if eventType == "message_updated" && msg != nil {
		if deleted, _ := msg.ContentAttributes["deleted"].(bool); deleted {
			return s.handleMessageUpdated(ctx, cfg, body)
		}
	}

	if eventType == "conversation_status_changed" && body.Conversation != nil {
		return s.handleConversationStatusChanged(ctx, cfg, body)
	}

	return nil
}

func (s *Service) handleOutgoingMessage(ctx context.Context, cfg *ChatwootConfig, body dto.ChatwootWebhookPayload) error {
	msg := body.GetMessage()
	if msg == nil || body.Conversation == nil {
		return nil
	}

	if strings.HasPrefix(msg.SourceID, "WAID:") {
		return nil
	}

	conv := body.Conversation
	chatJID := conv.ContactInbox.SourceID
	if chatJID == "" && conv.Meta.Sender.Identifier != "" {
		chatJID = conv.Meta.Sender.Identifier
	}
	if chatJID == "" && conv.Meta.Sender.PhoneNumber != "" {
		phone := strings.TrimPrefix(conv.Meta.Sender.PhoneNumber, "+")
		chatJID = phone + "@s.whatsapp.net"
	}
	if chatJID == "" {
		logger.Warn().Int("convID", conv.ID).Msg("[CW] no chat JID found for outgoing message")
		return nil
	}

	logger.Debug().Str("chatJID", chatJID).Str("content", msg.Content).Msg("[CW] sending outgoing message to WhatsApp")

	var replyTo *dto.ReplyContext
	if msg.ContentAttributes != nil {
		if inReplyTo, ok := msg.ContentAttributes["in_reply_to"].(float64); ok && inReplyTo > 0 {
			origMsg, err := s.msgRepo.FindByCWMessageID(ctx, cfg.SessionID, int(inReplyTo))
			if err == nil {
				replyTo = &dto.ReplyContext{MessageID: origMsg.ID}
			}
		}
	}

	if len(msg.Attachments) > 0 {
		for _, att := range msg.Attachments {
			attURL := att.DataURL
			if attURL == "" {
				attURL = att.URL
			}
			attURL = rewriteAttachmentURL(attURL, cfg.URL)
			if err := s.sendAttachmentToWhatsApp(ctx, cfg, chatJID, attURL, msg.Content, att.FileType, replyTo); err != nil {
				logger.Warn().Err(err).Msg("Failed to send attachment from Chatwoot to WhatsApp")
			}
		}
		return nil
	}

	content := msg.Content
	content = convertCWToWAMarkdown(content)

	_, err := s.messageSvc.SendText(ctx, cfg.SessionID, dto.SendTextReq{
		Phone:   chatJID,
		Body:    content,
		ReplyTo: replyTo,
	})
	return err
}

func (s *Service) handleMessageUpdated(ctx context.Context, cfg *ChatwootConfig, body dto.ChatwootWebhookPayload) error {
	webhookMsg := body.GetMessage()
	if webhookMsg == nil {
		return nil
	}

	cwMsgID := webhookMsg.ID
	storedMsg, err := s.msgRepo.FindByCWMessageID(ctx, cfg.SessionID, cwMsgID)
	if err != nil {
		return nil
	}

	_, _ = s.messageSvc.DeleteMessage(ctx, cfg.SessionID, dto.DeleteMessageReq{
		Phone:     storedMsg.ChatJID,
		MessageID: storedMsg.ID,
	})

	if body.Conversation != nil && body.Conversation.Status == "resolved" && !cfg.ReopenConversation {
		cacheKey := cfg.SessionID + "+" + body.Conversation.ContactInbox.SourceID
		s.convCache.Delete(cacheKey)
	}

	return nil
}

func (s *Service) handleConversationStatusChanged(ctx context.Context, cfg *ChatwootConfig, body dto.ChatwootWebhookPayload) error {
	if body.Conversation == nil {
		return nil
	}

	if body.Conversation.Status == "resolved" && !cfg.ReopenConversation {
		sourceID := body.Conversation.ContactInbox.SourceID
		if sourceID != "" {
			cacheKey := cfg.SessionID + "+" + sourceID
			s.convCache.Delete(cacheKey)
		}
	}

	return nil
}

func rewriteAttachmentURL(attURL, chatwootBaseURL string) string {
	if attURL == "" || chatwootBaseURL == "" {
		return attURL
	}

	parsed, err := url.Parse(attURL)
	if err != nil {
		return attURL
	}

	base, err := url.Parse(strings.TrimRight(chatwootBaseURL, "/"))
	if err != nil {
		return attURL
	}

	parsed.Scheme = base.Scheme
	parsed.Host = base.Host
	return parsed.String()
}

func filenameFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	segments := strings.Split(parsed.Path, "/")
	for i := len(segments) - 1; i >= 0; i-- {
		seg := segments[i]
		if seg != "" && strings.Contains(seg, ".") {
			decoded, err := url.PathUnescape(seg)
			if err != nil {
				return seg
			}
			return decoded
		}
	}
	return ""
}

func (s *Service) sendAttachmentToWhatsApp(ctx context.Context, cfg *ChatwootConfig, chatJID, attachmentURL, caption, fileType string, replyTo *dto.ReplyContext) error {
	resp, err := s.httpClient.Get(attachmentURL)
	if err != nil {
		return fmt.Errorf("download attachment: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read attachment: %w", err)
	}

	filename := filenameFromURL(attachmentURL)
	mimeType, _ := GetMIMETypeAndExt(filename, data)

	msgType := fileType
	if strings.Contains(mimeType, "/") {
		msgType = strings.Split(mimeType, "/")[0]
	}

	mediaReq := dto.SendMediaReq{
		Phone:    chatJID,
		Caption:  caption,
		MimeType: mimeType,
		FileName: filename,
		Base64:   base64.StdEncoding.EncodeToString(data),
		ReplyTo:  replyTo,
	}

	switch msgType {
	case "image":
		_, err = s.messageSvc.SendImage(ctx, cfg.SessionID, mediaReq)
	case "video":
		_, err = s.messageSvc.SendVideo(ctx, cfg.SessionID, mediaReq)
	case "audio":
		_, err = s.messageSvc.SendAudio(ctx, cfg.SessionID, mediaReq)
	default:
		_, err = s.messageSvc.SendDocument(ctx, cfg.SessionID, mediaReq)
	}

	return err
}
