## ADDED Requirements

### Requirement: validar conversa em cache antes de usar
Quando um ID de conversa é recuperado do cache, o sistema deve verificar se a conversa ainda existe no Chatwoot antes de usá-la.

#### Scenario: hit de cache com conversa ainda existente
- **WHEN** o cache retorna um `convID` e `GET /api/v1/accounts/{id}/conversations/{convID}` responde com sucesso
- **THEN** o `convID` em cache é retornado imediatamente; nenhuma recriação é feita

#### Scenario: hit de cache com conversa não encontrada (404 ou qualquer erro)
- **WHEN** o cache retorna um `convID` e a chamada `GetConversation` retorna erro (404, timeout, rede)
- **THEN** o cache para aquela chave (`sessionID+chatJID`) é invalidado
- **THEN** o slow path (`findOrCreateConversationSlowPath`) é executado para recriar a conversa
- **THEN** o novo `convID` resultante é retornado

#### Scenario: miss de cache
- **WHEN** o cache não tem entrada para `sessionID+chatJID`
- **THEN** comportamento atual mantido: singleflight + slow path; sem chamada `GetConversation`

### Requirement: novo método GetConversation no cliente Chatwoot
- **WHEN** `GetConversation(ctx, convID)` é chamado
- **THEN** realiza `GET /api/v1/accounts/{accountId}/conversations/{convID}` com o token de autenticação
- **THEN** retorna `(*Conversation, error)`; erro em qualquer status não-2xx
