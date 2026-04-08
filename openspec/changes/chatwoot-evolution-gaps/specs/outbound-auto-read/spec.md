## ADDED Requirements

### Requirement: marcar última mensagem WA como lida ao agente responder
Quando `cfg.MessageRead == true` e o agente envia uma mensagem outgoing, o sistema deve marcar a última mensagem recebida na conversa como lida no WhatsApp.

#### Scenario: envio outgoing bem-sucedido com MessageRead ativo
- **WHEN** `cfg.MessageRead == true` e uma mensagem outgoing é enviada com sucesso ao WA
- **THEN** o sistema busca a última mensagem recebida (fromMe=false) no chat; se encontrada, chama `messageSvc.MarkRead` com o ID e JID dessa mensagem

#### Scenario: nenhuma mensagem não lida disponível
- **WHEN** `cfg.MessageRead == true` mas não há mensagem com `fromMe=false` no banco para aquele chat
- **THEN** nenhuma chamada de mark-read é feita; sem erro

#### Scenario: mark-read falha
- **WHEN** `messageSvc.MarkRead` retorna erro
- **THEN** o erro é ignorado (best-effort); o envio da mensagem já foi bem-sucedido e não é revertido

#### Scenario: MessageRead desabilitado
- **WHEN** `cfg.MessageRead == false`
- **THEN** nenhuma chamada de mark-read é feita; comportamento atual mantido

### Requirement: nova flag messageRead na configuração por sessão
- **WHEN** a configuração Chatwoot de uma sessão é carregada
- **THEN** o campo `messageRead` (bool) está disponível; padrão `false`
