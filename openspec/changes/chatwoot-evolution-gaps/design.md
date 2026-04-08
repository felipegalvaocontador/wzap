## Context

A integração Chatwoot do wzap (`internal/integrations/chatwoot/`) é uma bridge bidirecional entre WhatsApp (via whatsmeow) e Chatwoot. Análise comparativa com a Evolution API identificou 10 gaps funcionais distribuídos em 4 categorias:

1. **Formatação inbound degradada**: `locationMessage`, `contactMessage`/`contactsArrayMessage` e `viewOnceMessageV2` chegam ao Chatwoot sem estrutura visual adequada.
2. **Sincronização de contato incompleta**: o campo `avatar_url` é enviado no nível errado (`additional_attributes`) e não é buscado no momento de criação do contato; o merge de contatos BR não usa a API de merge do Chatwoot.
3. **Features de UX ausentes no outbound**: assinatura de agente (`signMsg`) existe na config mas não é aplicada; auto-read e notificação de erro ao agente não estão implementados.
4. **Resiliência de cache**: IDs de conversa cacheados são usados sem verificar se a conversa ainda existe no Chatwoot.

Todos os gaps são contidos dentro de `internal/integrations/chatwoot/`. A única exceção é o campo `AvailableName` no DTO de webhook Chatwoot (`internal/dto/chatwoot.go`) necessário para `signMsg`.

## Goals / Non-Goals

**Goals:**
- Exibir `locationMessage` e `liveLocationMessage` com bloco estruturado (lat, lng, nome, endereço, URL Maps)
- Exibir `contactMessage`/`contactsArrayMessage` com template bold/italic e telefones numerados
- Enviar conteúdo real de `viewOnceMessageV2` (image/video/audio) com fallback textual
- Corrigir `avatar_url` para campo de primeiro nível em create/update de contato
- Buscar profilePicture WA no momento de criação de contato
- Implementar merge real de contatos BR via `POST /actions/contact_merge`
- Implementar `addLabelToContact` via SQL direto (opcional, requer `databaseUri`)
- Aplicar `signMsg` no outbound (texto + attachments)
- Implementar `messageRead` para marcar WA como lido ao agente responder
- Enviar mensagem privada ao agente quando envio WA falha
- Validar conversa em cache via `GET /conversations/{id}` antes de usar

**Non-Goals:**
- Não será implementado delay artificial humano no outbound
- Não serão tratados `listMessage` inbound, `interactiveMessage` (PIX), mensagens patrocinadas
- Não será criado contato individual de participantes de grupo
- Labels não são obrigatórias — ausência de `databaseUri` não bloqueia nada

## Decisions

### D1: `viewOnce` — desempacotar `vonce["message"]` no próprio handler
`handleViewOnce` recebe `vonce map[string]interface{}` (já é o objeto `viewOnceMessageV2`). O sub-mapa `vonce["message"]` contém o tipo real (imageMessage, videoMessage, audioMessage). A função `extractMediaInfo` existente em `parser.go` aceita qualquer `map[string]interface{}` — basta chamá-la com `vonce["message"]`. Fallback para texto placeholder se o sub-mapa não existir ou download falhar.

### D2: `avatar_url` — obtido via `wa.Manager` injetado no `Service`
O `Service` já tem `waMgr` (interface do wa Manager). A obtenção de foto de perfil (`GetProfilePicture(sessionID, jid)`) deve ser adicionada à interface `wa.Manager`. Se falhar ou retornar vazio, o contato é criado sem avatar (não-bloqueante).

### D3: Labels — arquivo separado `labels.go` com pgx direto
A função `addLabelToContact` usa SQL direto no banco do Chatwoot. Isolada em `labels.go` para não poluir `conversation.go`. Requer `cfg.DatabaseURI != ""`. Erros são logados (warn) e nunca bloqueiam o fluxo principal.

### D4: `signMsg` — aplicado após `convertCWToWAMarkdown`, antes de enviar
O nome do agente vem de `body.Conversation.Messages[0].Sender.AvailableName` (campo a adicionar ao DTO) com fallback para `body.Sender.Name`. Aplicado tanto em texto quanto como caption de attachments. Skip quando `senderName == ""` ou `!cfg.SignMsg`.

### D5: `autoRead` — nova flag por sessão (não global como na Evolution API)
A Evolution usa env var global. wzap adota a filosofia de config por sessão. Campo `MessageRead bool` em `ChatwootConfig`. A busca da última mensagem não lida usa `msgRepo.FindLastUnreadByChat(ctx, sessionID, chatJID)` — novo método no repo.

### D6: `GetConversation` — verificação apenas no hit de cache, não no slow path
O slow path já garante a consistência. A verificação é inserida apenas quando há cache hit em `findOrCreateConversation`. Se `GetConversation` falhar (qualquer erro), o cache é invalidado e o slow path é executado. Sem re-tentativa recursiva — o slow path já é idempotente.

### D7: `MergeContacts` — somente quando exatamente 2 variantes BR encontradas
Mantém a mesma precondição da Evolution API: `len(contacts) == 2 && cfg.MergeBRContacts && strings.HasPrefix(phone, "55")`. Base = contato com número de 14 chars (com `+` e com 9); mergee = 13 chars. Erro de merge é logado e ignorado; o contato com 14 chars é usado como resultado.

## Risks / Trade-offs

| Risco | Probabilidade | Mitigação |
|---|---|---|
| `viewOnce` com chave de mídia expirada | Média | Fallback para texto placeholder mantido |
| `GetConversation` adiciona latência por mensagem | Baixa | Aceitável; otimização com TTL de re-verificação em versão futura |
| SQL direto em tabelas do Chatwoot quebra em upgrades | Média | Condicionado a `databaseUri`; isolado em `labels.go`; erro não-bloqueante |
| `MergeContacts` falha com 422/500 do Chatwoot | Baixa | Erro ignorado; continua com contato com 9º dígito |
| `GetProfilePicture` lento no caminho crítico | Baixa | Contexto com timeout (padrão 5s); falha silenciosa |
| `MarkRead` sem mensagem não lida disponível | Média | `FindLastUnreadByChat` retorna nil sem erro; skip silencioso |
