package chatwoot

import (
	"context"
	"net/http"
	"sync"
	"time"

	"wzap/internal/dto"
	"wzap/internal/logger"
	"wzap/internal/model"
	"wzap/internal/repo"
)

type MessageService interface {
	SendText(ctx context.Context, sessionID string, req dto.SendTextReq) (string, error)
	SendImage(ctx context.Context, sessionID string, req dto.SendMediaReq) (string, error)
	SendVideo(ctx context.Context, sessionID string, req dto.SendMediaReq) (string, error)
	SendDocument(ctx context.Context, sessionID string, req dto.SendMediaReq) (string, error)
	SendAudio(ctx context.Context, sessionID string, req dto.SendMediaReq) (string, error)
	DeleteMessage(ctx context.Context, sessionID string, req dto.DeleteMessageReq) (string, error)
}
type JIDResolver interface {
	GetPNForLID(ctx context.Context, sessionID, lidJID string) string
}
type MediaDownloader interface {
	DownloadMediaByPath(ctx context.Context, sessionID, directPath string, encFileHash, fileHash, mediaKey []byte, fileLength int, mediaType string) ([]byte, error)
}

type Service struct {
	repo            Repo
	msgRepo         repo.MessageRepo
	clientFn        func(cfg *ChatwootConfig) CWClient
	messageSvc      MessageService
	convCache       sync.Map
	jidResolver     JIDResolver
	mediaDownloader MediaDownloader
	serverURL       string
	httpClient      *http.Client
}

func NewService(repo Repo, msgRepo repo.MessageRepo, messageSvc MessageService) *Service {
	return &Service{
		repo:    repo,
		msgRepo: msgRepo,
		clientFn: func(cfg *ChatwootConfig) CWClient {
			return NewClient(cfg.URL, cfg.AccountID, cfg.Token, &http.Client{Timeout: 30 * time.Second})
		},
		messageSvc: messageSvc,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *Service) SetJIDResolver(r JIDResolver)         { s.jidResolver = r }
func (s *Service) SetMediaDownloader(d MediaDownloader) { s.mediaDownloader = d }
func (s *Service) SetServerURL(url string)              { s.serverURL = url }

func (s *Service) OnEvent(sessionID string, event model.EventType, payload []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Debug().Str("session", sessionID).Str("event", string(event)).Msg("[CW] OnEvent received")

	cfg, err := s.repo.FindBySessionID(ctx, sessionID)
	if err != nil {
		logger.Debug().Str("session", sessionID).Err(err).Msg("[CW] config not found, skipping")
		return
	}
	if !cfg.Enabled {
		logger.Debug().Str("session", sessionID).Msg("[CW] integration disabled, skipping")
		return
	}

	switch event {
	case model.EventMessage:
		s.handleMessage(ctx, cfg, payload)
	case model.EventReceipt:
		s.handleReceipt(ctx, cfg, payload)
	case model.EventDeleteForMe:
		s.handleDelete(ctx, cfg, payload)
	case model.EventConnected:
		s.handleConnected(ctx, cfg, payload)
	case model.EventDisconnected:
		s.handleDisconnected(ctx, cfg, payload)
	case model.EventQR:
		s.handleQR(ctx, cfg, payload)
	case model.EventContact:
		s.handleContact(ctx, cfg, payload)
	case model.EventPushName:
		s.handlePushName(ctx, cfg, payload)
	case model.EventPicture:
		s.handlePicture(ctx, cfg, payload)
	}
}
