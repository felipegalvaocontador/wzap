## 1. Preparação: DTOs e interfaces

- [x] 1.1 Adicionar `AvatarURL string` em `CreateContactReq` e `UpdateContactReq` em `client.go`
- [x] 1.2 Adicionar campo `Sender` com `AvailableName string` na struct anônima de `Conversation.Messages[]` em `internal/dto/chatwoot.go`
- [x] 1.3 Adicionar método `GetConversation(ctx context.Context, convID int) (*Conversation, error)` à interface `CWClient` e implementar em `client.go`
- [x] 1.4 Adicionar método `MergeContacts(ctx context.Context, baseID, mergeeID int) error` à interface `CWClient` e implementar em `client.go`
- [x] 1.5 Adicionar campos `MessageRead bool` e `DatabaseURI string` em `ChatwootConfig` (`config.go`)
- [x] 1.6 Adicionar método `GetProfilePicture(ctx context.Context, sessionID, jid string) (string, error)` à interface `wa.Manager` e implementar

## 2. Formatação inbound: location e contact

- [x] 2.1 Atualizar bloco `locationMessage` em `parser.go`: usar template `*Localização:*\n\n_Latitude:_ {lat}\n_Longitude:_ {lng}\n[_Nome:_ {name}\n][_Endereço:_ {address}\n]_URL:_ https://www.google.com/maps/search/?api=1&query={lat},{lng}`
- [x] 2.2 Atualizar bloco `liveLocationMessage` em `inbound_message.go` (`handleLiveLocation`): aplicar o mesmo template que `locationMessage` (lat, lng, nome, endereço, URL)
- [x] 2.3 Atualizar `formatVCard` em `parser.go`: substituir template para `*Contato:*\n\n_Nome:_ {nome}\n_Número (1):_ {tel1}\n_Número (2):_ {tel2}...`
- [x] 2.4 Atualizar bloco `contactsArrayMessage` em `parser.go`: usar `displayName` do objeto como nome (fallback para `FN` do vCard); remover prefixo global `📇 Contatos:`; cada item inicia com `*Contato:*`

## 3. viewOnce com conteúdo real

- [x] 3.1 Atualizar `handleViewOnce` em `inbound_message.go`: desempacotar `vonce["message"]`; detectar tipo interno (`imageMessage`, `videoMessage`, `audioMessage`); chamar `extractMediaInfo` sobre o sub-mapa e baixar via `s.mediaDownloader`
- [x] 3.2 Enviar arquivo baixado como attachment via `client.CreateMessageWithAttachment`; manter fallback para texto `[mensagem vista uma vez]` quando download falhar ou sub-mapa ausente

## 4. Sincronização de avatar de contato

- [x] 4.1 Corrigir `handlePicture` em `inbound_events.go`: substituir `AdditionalAttributes: map[string]any{"avatar_url": data.URL}` por `AvatarURL: data.URL` em `UpdateContactReq`
- [x] 4.2 Em `findOrCreateConversationSlowPath` (`conversation.go`): ao criar contato novo, chamar `s.waMgr.GetProfilePicture(ctx, cfg.SessionID, jid)` e popular `CreateContactReq.AvatarURL` com o resultado (falha silenciosa)

## 5. Merge de contatos BR

- [x] 5.1 Em `findOrCreateConversationSlowPath` (`conversation.go`): após coletar contatos com variante BR, quando `len(contacts) == 2 && cfg.MergeBRContacts`, identificar base (14 chars) e mergee (13 chars), chamar `client.MergeContacts(ctx, baseID, mergeeID)`; usar base como `contactID`; erro de merge logado e ignorado

## 6. Labels de inbox em contatos

- [x] 6.1 Criar arquivo `internal/integrations/chatwoot/labels.go` com função `addLabelToContact(ctx context.Context, dbURI, inboxName string, contactID int) error` usando as queries SQL: upsert em `tags` + insert em `taggings` (verificar existência antes)
- [x] 6.2 Em `findOrCreateConversationSlowPath` (`conversation.go`): após criar contato com sucesso, chamar `addLabelToContact(ctx, cfg.DatabaseURI, cfg.InboxName, contact.ID)` condicionado a `cfg.DatabaseURI != ""`

## 7. Assinatura de agente no outbound (signMsg)

- [x] 7.1 Em `outbound.go` (`handleOutgoingMessage`): após `convertCWToWAMarkdown`, inserir lógica de sign: obter `senderName` de `body.Conversation.Messages[0].Sender.AvailableName` (fallback para `body.Sender.Name`); se `cfg.SignMsg && senderName != ""`, formatar `*{senderName}:*{delimiter}{content}`
- [x] 7.2 Aplicar o mesmo `content` assinado como `caption` nos attachments (linha de `sendAttachmentToWhatsApp`)
- [x] 7.3 Transformar `cfg.SignDelimiter`: substituir `\n` literal por newline real (`strings.ReplaceAll(d, `\n`, "\n")`)

## 8. Auto-read (messageRead)

- [x] 8.1 Adicionar método `FindLastUnreadByChat(ctx context.Context, sessionID, chatJID string) (*model.Message, error)` em `internal/repo/` (retorna última mensagem com `from_me=false` para o chat)
- [x] 8.2 Adicionar método `MarkRead` à interface `MessageService` em `internal/integrations/chatwoot/service.go` se ainda não presente
- [x] 8.3 Em `outbound.go`: após envio bem-sucedido de qualquer mensagem (texto ou attachment), se `cfg.MessageRead`, buscar via `s.msgRepo.FindLastUnreadByChat` e chamar `s.messageSvc.MarkRead`; erros ignorados (best-effort)

## 9. Notificação de erro ao agente

- [x] 9.1 Em `outbound.go`: criar método helper `sendErrorToAgent(ctx, cfg, convID, err)` que cria mensagem privada no Chatwoot com `message_type: "outgoing"`, `private: true`, conteúdo `_Mensagem não enviada: {err}_`
- [x] 9.2 Chamar `sendErrorToAgent` após falha em `messageSvc.SendText`
- [x] 9.3 Chamar `sendErrorToAgent` após falha em `sendAttachmentToWhatsApp` (dentro do loop de attachments)

## 10. Validação de conversa em cache

- [x] 10.1 Em `findOrCreateConversation` (`conversation.go`): após hit de cache, chamar `client.GetConversation(ctx, convID)`; se erro, chamar `s.cache.DeleteConv(ctx, cfg.SessionID, chatJID)` e prosseguir para o slow path
- [x] 10.2 Aplicar a mesma verificação no segundo check de cache dentro do `convFlight.Do`

## 11. Testes e validação

- [x] 11.1 Escrever testes unitários para os novos formatadores em `parser.go` (location rico, contact markdown, contactsArray com displayName)
- [x] 11.2 Escrever testes para `handleViewOnce` (caso com download bem-sucedido, caso de fallback)
- [x] 11.3 Escrever testes para `sendErrorToAgent` (falha de texto, falha de attachment)
- [x] 11.4 Escrever testes para `GetConversation` e `MergeContacts` no `client.go`
- [x] 11.5 Escrever testes para `addLabelToContact` (com e sem DatabaseURI, label já existente)
- [x] 11.6 Executar `go test -race ./internal/integrations/chatwoot/... ./internal/dto/...` e garantir PASS
- [x] 11.7 Executar `golangci-lint run ./internal/integrations/chatwoot/... ./internal/dto/...` e garantir 0 issues
