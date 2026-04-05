package chatwoot

import (
	"context"
	"strings"

	"wzap/internal/logger"
)

func (s *Service) handleMessage(ctx context.Context, cfg *ChatwootConfig, payload []byte) {
	data, err := parseMessagePayload(payload)
	if err != nil {
		logger.Warn().Err(err).Msg("[CW] Failed to parse message payload")
		return
	}

	logger.Debug().Str("chat", data.Info.Chat).Str("id", data.Info.ID).Bool("fromMe", data.Info.IsFromMe).Msg("[CW] handleMessage")

	chatJID := data.Info.Chat
	if chatJID == "" {
		logger.Warn().Msg("[CW] chatJID empty, skipping")
		return
	}

	if strings.HasSuffix(chatJID, "@lid") && s.jidResolver != nil {
		if pn := s.jidResolver.GetPNForLID(ctx, cfg.SessionID, chatJID); pn != "" {
			logger.Debug().Str("lid", chatJID).Str("pn", pn).Msg("[CW] resolved LID to PN")
			chatJID = pn + "@s.whatsapp.net"
		}
	}

	if shouldIgnoreJID(chatJID, cfg.IgnoreGroups, cfg.IgnoreJIDs) {
		logger.Debug().Str("chat", chatJID).Msg("[CW] JID ignored, skipping")
		return
	}

	pushName := data.Info.PushName
	fromMe := data.Info.IsFromMe
	msgID := data.Info.ID

	convID, err := s.findOrCreateConversation(ctx, cfg, chatJID, pushName)
	if err != nil {
		logger.Warn().Err(err).Str("session", cfg.SessionID).Str("chatJID", chatJID).Msg("Failed to find or create Chatwoot conversation")
		return
	}

	client := s.clientFn(cfg)

	mediaInfo := extractMediaInfo(data.Message)
	if mediaInfo != nil {
		s.handleMediaMessage(ctx, cfg, convID, msgID, fromMe, data.Message)
		return
	}

	text := extractTextFromMessage(data.Message)
	logger.Debug().Str("text", text).Interface("msg", data.Message).Msg("[CW] extracted text")
	text = convertWAToCWMarkdown(text)

	if !fromMe && data.Info.IsGroup && cfg.SignMsg && pushName != "" {
		phone := extractPhone(chatJID)
		text = formatGroupContent(phone, pushName, text, fromMe)
	}

	if text != "" {
		messageType := "outgoing"
		if !fromMe {
			messageType = "incoming"
		}

		msgReq := MessageReq{
			Content:     text,
			MessageType: messageType,
			SourceID:    "WAID:" + msgID,
		}

		stanzaID := extractStanzaID(data.Message)
		if stanzaID != "" {
			origMsg, err := s.msgRepo.FindByID(ctx, cfg.SessionID, stanzaID)
			if err == nil && origMsg.CWMessageID != nil && *origMsg.CWMessageID != 0 {
				msgReq.InReplyTo = *origMsg.CWMessageID
			}
		}

		msg, err := client.CreateMessage(ctx, convID, msgReq)
		if err != nil {
			logger.Warn().Err(err).Str("session", cfg.SessionID).Msg("Failed to create Chatwoot message")
			return
		}
		if msgID != "" {
			_ = s.msgRepo.UpdateChatwootRef(ctx, cfg.SessionID, msgID, msg.ID, convID, msg.SourceID)
		}
	}
}

func (s *Service) handleMediaMessage(ctx context.Context, cfg *ChatwootConfig, convID int, msgID string, fromMe bool, msg map[string]interface{}) {
	info := extractMediaInfo(msg)
	if info == nil {
		logger.Warn().Msg("[CW] no media info found in message")
		return
	}

	if s.mediaDownloader == nil {
		logger.Warn().Msg("[CW] media downloader not configured, cannot download WhatsApp media")
		return
	}

	data, err := s.mediaDownloader.DownloadMediaByPath(ctx, cfg.SessionID, info.DirectPath, info.FileEncSHA256, info.FileSHA256, info.MediaKey, info.FileLength, info.MediaType)
	if err != nil {
		logger.Warn().Err(err).Str("mediaType", info.MediaType).Msg("[CW] failed to download WhatsApp media")
		return
	}

	mimeType := info.MimeType
	if mimeType == "" {
		mimeType, _ = GetMIMETypeAndExt("", data)
	}

	filename := info.FileName
	if filename == "" {
		ext := mimeTypeToExt(mimeType)
		filename = info.MediaType + ext
	}

	caption := extractTextFromMessage(msg)
	client := s.clientFn(cfg)
	messageType := "incoming"
	if fromMe {
		messageType = "outgoing"
	}
	sourceID := ""
	if msgID != "" {
		sourceID = "WAID:" + msgID
	}
	cwMsg, err := client.CreateMessageWithAttachment(ctx, convID, caption, filename, data, mimeType, messageType, sourceID)
	if err != nil {
		logger.Warn().Err(err).Msg("[CW] failed to upload media to Chatwoot")
		return
	}

	if msgID != "" {
		_ = s.msgRepo.UpdateChatwootRef(ctx, cfg.SessionID, msgID, cwMsg.ID, convID, cwMsg.SourceID)
	}
}
