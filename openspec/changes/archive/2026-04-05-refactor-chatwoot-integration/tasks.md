## 1. Criar service.go slim

- [x] 1.1 Criar `service.go` novo contendo apenas: interfaces (MessageService, JIDResolver, MediaDownloader), struct Service, NewService, SetJIDResolver, SetMediaDownloader, SetServerURL e OnEvent
- [x] 1.2 Verificar que `service.go` tem menos de 100 linhas

## 2. Extrair inbound handlers (WA→CW)

- [x] 2.1 Criar `inbound_message.go` com handleMessage e handleMediaMessage (movidos de service.go e sync.go)
- [x] 2.2 Criar `inbound_events.go` com handleReceipt, handleDelete, handleConnected, handleDisconnected, handleQR, handleContact, handlePushName, handlePicture (movidos de service.go)

## 3. Extrair conversation lifecycle

- [x] 3.1 Criar `conversation.go` com findOrCreateConversation, findOrCreateBotConversation, webhookURL e Configure (movidos de sync.go)

## 4. Renomear outbound

- [x] 4.1 Renomear `webhook.go` para `outbound.go` (manter todo o conteúdo: HandleIncomingWebhook, handleOutgoingMessage, handleMessageUpdated, handleConversationStatusChanged, rewriteAttachmentURL, sendAttachmentToWhatsApp)

## 5. Limpeza

- [x] 5.1 Remover `sync.go` (todo conteúdo já redistribuído)
- [x] 5.2 Remover handlers e métodos antigos de `service.go` original

## 6. Validação

- [x] 6.1 Executar `go build ./...` e confirmar compilação sem erros
- [x] 6.2 Executar `go test -race ./internal/integrations/chatwoot/...` e confirmar todos os testes passando
