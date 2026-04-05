package chatwoot

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"wzap/internal/logger"
)

func (s *Service) findOrCreateConversation(ctx context.Context, cfg *ChatwootConfig, chatJID, pushName string) (int, error) {
	cacheKey := cfg.SessionID + "+" + chatJID

	val, _ := s.convCache.LoadOrStore(cacheKey, &sync.Mutex{})
	mu := val.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	client := s.clientFn(cfg)
	phone := extractPhone(chatJID)

	var contacts []Contact
	if cfg.MergeBRContacts && strings.HasPrefix(phone, "55") {
		contacts1, _ := client.FilterContacts(ctx, phone)
		phoneVariant := addOrRemoveBR9thDigit(phone)
		contacts2, _ := client.FilterContacts(ctx, phoneVariant)
		contacts = append(contacts1, contacts2...)
	} else {
		var err error
		contacts, err = client.FilterContacts(ctx, phone)
		if err != nil {
			contacts = nil
		}
	}

	var contactID int
	if len(contacts) == 0 {
		name := pushName
		if name == "" {
			name = phone
		}
		contact, err := client.CreateContact(ctx, CreateContactReq{
			InboxID:     cfg.InboxID,
			Name:        name,
			Identifier:  chatJID,
			PhoneNumber: "+" + phone,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to create contact: %w", err)
		}
		logger.Debug().Int("contactID", contact.ID).Str("phone", phone).Msg("[CW] contact created")
		contactID = contact.ID
	} else {
		contactID = contacts[0].ID
		logger.Debug().Int("contactID", contactID).Str("phone", phone).Msg("[CW] contact found")
		if pushName != "" && contacts[0].Name != pushName {
			_ = client.UpdateContact(ctx, contactID, UpdateContactReq{Name: pushName})
		}
	}

	conversations, err := client.ListContactConversations(ctx, contactID)
	if err != nil {
		return 0, fmt.Errorf("failed to list conversations: %w", err)
	}

	for _, conv := range conversations {
		if conv.InboxID == cfg.InboxID {
			if conv.Status == "resolved" && cfg.ReopenConversation {
				reopenStatus := "open"
				if cfg.ConversationPending {
					reopenStatus = "pending"
				}
				if err := client.UpdateConversationStatus(ctx, conv.ID, reopenStatus); err != nil {
					logger.Warn().Err(err).Int("convID", conv.ID).Msg("Failed to reopen conversation")
				}
				return conv.ID, nil
			}
			if conv.Status != "resolved" {
				return conv.ID, nil
			}
		}
	}

	req := CreateConversationReq{
		InboxID:   cfg.InboxID,
		SourceID:  chatJID,
		ContactID: contactID,
	}
	if cfg.ConversationPending {
		req.Status = "pending"
	}

	conv, err := client.CreateConversation(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conv.ID, nil
}

func (s *Service) findOrCreateBotConversation(ctx context.Context, cfg *ChatwootConfig) (int, error) {
	client := s.clientFn(cfg)

	contacts, err := client.FilterContacts(ctx, cfg.SessionID)
	if err != nil {
		return 0, err
	}

	var contactID int
	if len(contacts) == 0 {
		contact, err := client.CreateContact(ctx, CreateContactReq{
			InboxID:    cfg.InboxID,
			Name:       cfg.SessionID,
			Identifier: cfg.SessionID,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to create bot contact: %w", err)
		}
		contactID = contact.ID
	} else {
		contactID = contacts[0].ID
	}

	conversations, err := client.ListContactConversations(ctx, contactID)
	if err != nil {
		return 0, err
	}

	for _, conv := range conversations {
		if conv.InboxID == cfg.InboxID && conv.Status != "resolved" {
			return conv.ID, nil
		}
	}

	conv, err := client.CreateConversation(ctx, CreateConversationReq{
		InboxID:   cfg.InboxID,
		ContactID: contactID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create bot conversation: %w", err)
	}

	return conv.ID, nil
}

func (s *Service) webhookURL(sessionID string) string {
	base := s.serverURL
	if base == "" {
		base = "http://localhost:8080"
	}
	return fmt.Sprintf("%s/chatwoot/webhook/%s", base, sessionID)
}

func (s *Service) Configure(ctx context.Context, cfg *ChatwootConfig) error {
	if cfg.InboxName == "" {
		cfg.InboxName = "wzap"
	}

	client := s.clientFn(cfg)
	whURL := s.webhookURL(cfg.SessionID)

	inboxes, err := client.ListInboxes(ctx)
	if err != nil {
		inboxes = nil
	}

	found := false
	for _, inbox := range inboxes {
		if inbox.ID == cfg.InboxID {
			found = true
			_ = client.UpdateInboxWebhook(ctx, cfg.InboxID, whURL)
			break
		}
	}

	if !found && cfg.InboxID == 0 {
		inbox, err := client.CreateInbox(ctx, cfg.InboxName, whURL)
		if err != nil {
			return fmt.Errorf("failed to auto-create inbox: %w", err)
		}
		cfg.InboxID = inbox.ID
	}

	cfg.Enabled = true
	return s.repo.Upsert(ctx, cfg)
}
