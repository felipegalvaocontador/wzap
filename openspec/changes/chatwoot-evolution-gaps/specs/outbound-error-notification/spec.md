## ADDED Requirements

### Requirement: mensagem privada ao agente quando envio WA falha
Quando o sistema falha ao enviar uma mensagem ao WhatsApp (texto ou attachment), deve criar uma mensagem privada na conversa do Chatwoot informando o agente sobre a falha.

#### Scenario: falha ao enviar texto ao WA
- **WHEN** `messageSvc.SendText` retorna erro
- **THEN** uma mensagem é criada no Chatwoot na mesma conversa com `message_type: "outgoing"`, `private: true`, e conteúdo `_Mensagem não enviada: {erro}_`

#### Scenario: falha ao enviar attachment ao WA
- **WHEN** `sendAttachmentToWhatsApp` retorna erro para um attachment
- **THEN** uma mensagem privada é criada no Chatwoot com o erro do attachment; o loop de attachments continua para os demais

#### Scenario: envio bem-sucedido
- **WHEN** `messageSvc.SendText` ou `sendAttachmentToWhatsApp` retorna sem erro
- **THEN** nenhuma mensagem de erro é criada

#### Scenario: criação da mensagem de erro falha
- **WHEN** `client.CreateMessage` para a notificação de erro retorna erro
- **THEN** o erro secundário é silenciosamente ignorado; o fluxo não é afetado

### Requirement: mensagem de erro usa private: true
- **WHEN** a notificação de falha é enviada ao Chatwoot
- **THEN** `private: true` é definido no payload, tornando a mensagem visível apenas para agentes (não para o contato WA)
