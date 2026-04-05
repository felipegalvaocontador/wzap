## ADDED Requirements

### Requirement: Service file contains only orchestration concerns
O arquivo `service.go` SHALL conter apenas a definição de interfaces, struct `Service`, constructor `NewService`, setters de dependências e o método `OnEvent` (router de eventos). Nenhum handler de evento individual SHALL existir em `service.go`.

#### Scenario: service.go contém apenas struct e router
- **WHEN** o arquivo `service.go` é inspecionado
- **THEN** ele contém apenas: type interfaces (MessageService, JIDResolver, MediaDownloader), type Service struct, func NewService, funcs Set* (setters), func OnEvent

#### Scenario: service.go tem menos de 100 linhas
- **WHEN** o arquivo `service.go` é contabilizado
- **THEN** ele possui menos de 100 linhas de código

### Requirement: Inbound message handlers em arquivo dedicado
Os métodos `handleMessage` e `handleMediaMessage` SHALL existir em `inbound_message.go`.

#### Scenario: handleMessage está em inbound_message.go
- **WHEN** o projeto compila com sucesso
- **THEN** os métodos `handleMessage` e `handleMediaMessage` do receiver `*Service` estão definidos em `inbound_message.go`

### Requirement: Inbound event handlers em arquivo dedicado
Os métodos `handleReceipt`, `handleDelete`, `handleConnected`, `handleDisconnected`, `handleQR`, `handleContact`, `handlePushName` e `handlePicture` SHALL existir em `inbound_events.go`.

#### Scenario: Event handlers agrupados em inbound_events.go
- **WHEN** o projeto compila com sucesso
- **THEN** todos os 8 handlers de eventos WA (exceto handleMessage e handleMediaMessage) estão definidos em `inbound_events.go`

### Requirement: Outbound handlers renomeados de webhook.go
O arquivo `webhook.go` SHALL ser renomeado para `outbound.go`. Todos os métodos existentes (`HandleIncomingWebhook`, `handleOutgoingMessage`, `handleMessageUpdated`, `handleConversationStatusChanged`, `rewriteAttachmentURL`, `sendAttachmentToWhatsApp`) SHALL permanecer neste arquivo.

#### Scenario: webhook.go não existe mais
- **WHEN** o diretório `internal/integrations/chatwoot/` é listado
- **THEN** `webhook.go` não existe e `outbound.go` contém todos os handlers Chatwoot→WhatsApp

### Requirement: Conversation lifecycle em arquivo dedicado
Os métodos `findOrCreateConversation`, `findOrCreateBotConversation`, `webhookURL` e `Configure` SHALL existir em `conversation.go`.

#### Scenario: sync.go não existe mais
- **WHEN** o diretório `internal/integrations/chatwoot/` é listado
- **THEN** `sync.go` não existe e `conversation.go` contém toda a lógica de conversation lifecycle

### Requirement: Zero mudança de comportamento
A refatoração SHALL preservar exatamente o mesmo comportamento. Nenhuma signature de método, tipo exported, ou lógica de negócio SHALL mudar.

#### Scenario: Testes passam sem alteração
- **WHEN** `go test ./internal/integrations/chatwoot/...` é executado
- **THEN** todos os testes existentes passam sem nenhuma modificação nos arquivos de teste

#### Scenario: Build compila com sucesso
- **WHEN** `go build ./...` é executado
- **THEN** a compilação é bem-sucedida sem erros
