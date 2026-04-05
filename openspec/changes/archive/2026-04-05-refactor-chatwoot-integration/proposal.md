## Why

O pacote `internal/integrations/chatwoot/` cresceu organicamente durante a implementação de features (media download, webhook bidirection, conversation lifecycle). O resultado é um `service.go` com 391 linhas que mistura interfaces, event routing, e 9 handlers distintos; um `sync.go` com nome confuso que agrupa media handling, conversation management e configuração; e nenhuma separação clara entre o fluxo WhatsApp→Chatwoot e Chatwoot→WhatsApp. A refatoração agora evita que o custo de manutenção cresça exponencialmente com as próximas features.

## What Changes

- Dividir `service.go` (391 lines) em arquivo slim (~70 lines) com struct/constructor/OnEvent + arquivos dedicados por responsabilidade
- Renomear e decompor `sync.go` (233 lines): media handling vai para `inbound_message.go`, conversation lifecycle vai para `conversation.go`
- Renomear `webhook.go` para `outbound.go` para explicitar a direção do fluxo (Chatwoot→WhatsApp)
- Criar `inbound_message.go` (handleMessage + handleMediaMessage) e `inbound_events.go` (handleReceipt, handleDelete, handleConnected, handleDisconnected, handleQR, handleContact, handlePushName, handlePicture)
- Criar `conversation.go` (findOrCreateConversation, findOrCreateBotConversation, webhookURL, Configure)
- Nenhuma mudança de comportamento ou API pública — refatoração pura de organização interna

## Capabilities

### New Capabilities
- `chatwoot-file-organization`: Reorganização dos arquivos do pacote chatwoot em estrutura flat com nomes que comunicam direção de fluxo e responsabilidade

### Modified Capabilities

## Impact

- Apenas arquivos em `internal/integrations/chatwoot/` são afetados
- Nenhuma API HTTP muda
- Nenhuma interface pública muda (mesmo pacote, sem exports novos)
- Testes existentes continuam funcionando sem mudanças (métodos permanecem no mesmo receiver `*Service`)
- Zero impacto em dependências externas
