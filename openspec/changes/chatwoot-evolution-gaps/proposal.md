## Why

Análise comparativa com a Evolution API revelou 10 gaps funcionais na integração Chatwoot do wzap: tipos de mensagem exibidos de forma degradada (location, contato, viewOnce), sincronização incompleta de avatar/contatos, e ausência de features de UX esperadas por operadores (assinatura de agente, auto-read, notificação de erro). Fechar esses gaps coloca o wzap à frente da Evolution API em qualidade da integração.

## What Changes

- **Formatação de mensagem**: `locationMessage` e `liveLocationMessage` passam a exibir bloco estruturado com latitude, longitude, endereço e link Google Maps; `contactMessage`/`contactsArrayMessage` formatam vCard com bold/italic Chatwoot markdown; `viewOnceMessageV2` baixa e envia o arquivo real em vez de placeholder textual.
- **Sincronização de contato**: `CreateContactReq` e `UpdateContactReq` passam a incluir `avatar_url` como campo de primeiro nível (corrigindo o uso incorreto de `additional_attributes.avatar_url`); o `findOrCreateConversationSlowPath` busca a foto de perfil WA antes de criar o contato.
- **Merge de contatos BR**: quando dois registros com/sem 9º dígito são encontrados, wzap passa a chamar `POST /api/v1/accounts/{id}/actions/contact_merge` para fundir os registros no Chatwoot.
- **Labels em contatos**: ao criar um contato, wzap pode opcionalmente adicionar a label com o nome do inbox via SQL direto no banco do Chatwoot (requer `databaseUri` configurado).
- **Assinatura de agente no outbound**: quando `signMsg: true`, cada mensagem enviada pelo agente recebe prefixo `*NomeDoCAgente:*` + delimitador configurável.
- **Auto-read**: quando `messageRead: true`, ao responder via Chatwoot o wzap marca a última mensagem recebida no WhatsApp como lida.
- **Notificação de erro ao agente**: quando o envio de mensagem ao WhatsApp falha, wzap cria mensagem privada (`private: true`) na conversa do Chatwoot informando o agente.
- **Validação de conversa em cache**: ao usar um ID de conversa do cache, wzap verifica se ela ainda existe no Chatwoot; se não, invalida o cache e recria.

## Capabilities

### New Capabilities

- `location-message-rich`: formatação estruturada de locationMessage/liveLocationMessage com lat, lng, nome, endereço e URL Google Maps
- `contact-message-rich`: formatação de contactMessage/contactsArrayMessage com template bold/italic e exibição de telefones numerados
- `view-once-media-download`: download e envio do arquivo real de viewOnceMessageV2 (image/video/audio) com fallback textual
- `contact-avatar-sync`: campo `avatar_url` de primeiro nível em CreateContact/UpdateContact + busca de profilePicture ao criar contato
- `contact-br-merge`: chamada à API de merge do Chatwoot para unificar contatos BR com/sem 9º dígito
- `contact-inbox-labels`: label com nome do inbox adicionada a contatos via SQL direto no Chatwoot (opcional, requer `databaseUri`)
- `outbound-sign-message`: assinatura `*AgenteName:*` prefixada às mensagens outbound quando `signMsg: true`
- `outbound-auto-read`: mark-read no WhatsApp ao agente responder, quando `messageRead: true`
- `outbound-error-notification`: mensagem privada ao agente na conversa quando envio WA falha
- `conversation-cache-validation`: verificação de existência de conversa cacheada via `GET /conversations/{id}` antes de usar o ID

### Modified Capabilities

## Impact

**Arquivos modificados (sem criar novos):**
- `internal/integrations/chatwoot/parser.go` — location, contact formatting
- `internal/integrations/chatwoot/inbound_message.go` — handleViewOnce
- `internal/integrations/chatwoot/client.go` — avatar_url nos req, GetConversation, MergeContacts
- `internal/integrations/chatwoot/conversation.go` — busca avatar, merge BR, validação de cache
- `internal/integrations/chatwoot/outbound.go` — signMsg, sendErrorToAgent, auto-read
- `internal/integrations/chatwoot/config.go` — MessageRead, DatabaseURI
- `internal/integrations/chatwoot/inbound_events.go` — corrigir campo avatar_url
- `internal/dto/chatwoot.go` — AvailableName em Conversation.Messages[].Sender

**Arquivos novos:**
- `internal/integrations/chatwoot/labels.go` — addLabelToContact via SQL

**Dependências externas:** nenhuma nova. Labels via SQL usa `pgx` já presente no projeto.

**Não-objetivos:**
- Não serão implementados: listMessage inbound, interactiveMessage (PIX), mensagens patrocinadas (Ads), criação de contato individual de participantes de grupo, delay artificial humano no outbound
- Labels são opcionais — sistema continua funcional sem `databaseUri` configurado
- Merge BR só ocorre quando exatamente 2 variantes são encontradas (mesma lógica da Evolution API)

**Riscos e mitigações:**
- `viewOnce` download pode falhar por expiração de chave de mídia → fallback para texto placeholder mantido
- `GetConversation` a cada mensagem adiciona uma chamada HTTP → aceitável; pode ser otimizado com TTL de re-validação em versão futura
- SQL direto no banco do Chatwoot (labels) é frágil a mudanças de schema → condicionado a `databaseUri` e isolado em função separada; erro não bloqueia o fluxo principal
- `MergeContacts` 422/500 pode ocorrer se Chatwoot não suportar merge → erro ignorado, continua com o contato encontrado
