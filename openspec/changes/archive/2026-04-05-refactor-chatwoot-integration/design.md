## Context

O pacote `internal/integrations/chatwoot/` tem 8 arquivos Go (excluindo testes) que totalizam ~1650 linhas. O `service.go` acumulou 391 linhas com 9 handlers de eventos WhatsApp, interfaces, constructor e event router. O `sync.go` (233 lines) mistura media download, conversation lifecycle e configuração de inbox. O `webhook.go` (211 lines) processa webhooks Chatwoot→WhatsApp mas o nome não comunica essa direção.

Estado atual dos arquivos:
- `service.go` (391 lines) — interfaces + struct + OnEvent + 9 handlers WA→CW
- `sync.go` (233 lines) — handleMediaMessage + findOrCreateConversation + Configure
- `webhook.go` (211 lines) — HandleIncomingWebhook + handlers CW→WA
- `parser.go` (267 lines) — parsing de payloads WhatsApp
- `client.go` (346 lines) — HTTP client API Chatwoot
- `handler.go` (240 lines) — Fiber HTTP endpoints
- `media.go` (105 lines) — MIME type helpers
- `jid.go` (66 lines) — JID/phone helpers
- `config.go`, `repo.go` — modelos e database

## Goals / Non-Goals

**Goals:**
- Cada arquivo com responsabilidade clara e < 200 linhas
- Nomes de arquivo que comunicam a direção do fluxo (inbound = WA→CW, outbound = CW→WA)
- Conversation lifecycle isolado em arquivo dedicado
- Zero mudança de comportamento — testes passam sem alteração

**Non-Goals:**
- Criar subpastas (`inbound/`, `outbound/`) — o pacote não é grande o suficiente para justificar
- Refatorar lógica interna dos métodos (ex: quebrar `findOrCreateConversation` em submétodos)
- Alterar interfaces públicas, signatures de métodos, ou nomes de tipos exported
- Refatorar `client.go`, `handler.go`, `parser.go`, `config.go`, `repo.go` — já estão bem organizados

## Decisions

### 1. Estrutura flat com prefixos `inbound_` / `outbound_`

**Decisão:** Usar prefixos nos nomes de arquivo em vez de subpastas.

**Alternativa considerada:** Subpastas `inbound/`, `outbound/`, `conversation/`. Descartada porque exigiria criar novos pacotes Go, mover tipos entre pacotes, e lidar com dependências circulares — complexidade desproporcional para ~1650 linhas.

**Resultado:**
```
service.go           → service.go (slim: struct, constructor, interfaces, OnEvent)
service.go handlers  → inbound_message.go (handleMessage, handleMediaMessage)
service.go handlers  → inbound_events.go (handleReceipt..handlePicture)
sync.go              → conversation.go (findOrCreate*, webhookURL, Configure)
webhook.go           → outbound.go (HandleIncomingWebhook, handleOutgoing*, sendAttachment*)
```

### 2. `sync.go` desaparece completamente

**Decisão:** Eliminar `sync.go` redistribuindo:
- `handleMediaMessage` → `inbound_message.go` (junto com `handleMessage`, que é quem o chama)
- `findOrCreateConversation`, `findOrCreateBotConversation`, `webhookURL`, `Configure` → `conversation.go`

**Rationale:** O nome "sync" não comunica nada útil. As funções dentro dele pertencem a dois domínios distintos (message handling vs conversation lifecycle).

### 3. Manter todos os métodos no receiver `*Service`

**Decisão:** Não criar novos types ou structs. Todos os métodos continuam em `*Service`.

**Rationale:** Como tudo está no mesmo pacote Go, mover métodos entre arquivos é transparente para o compilador. Criar types separados forçaria refatorar a injeção de dependências.

## Risks / Trade-offs

- **[Risk] Git blame perde histórico** → Usar `git mv` não aplica aqui pois é redistribuição de conteúdo. Mitigação: commit message claro referenciando a refatoração.
- **[Risk] Merge conflicts com branches ativas** → Mitigação: fazer tudo em um único commit atômico.
- **[Trade-off] Arquivos `inbound_events.go` pode ficar grande** → Aceitável porque são handlers simples e independentes (~200 lines). Se crescer, pode ser dividido depois.
