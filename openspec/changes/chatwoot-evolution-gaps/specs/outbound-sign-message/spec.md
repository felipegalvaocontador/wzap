## ADDED Requirements

### Requirement: mensagens outbound prefixadas com nome do agente
Quando `cfg.SignMsg == true` e o nome do agente está disponível no payload do webhook, todas as mensagens enviadas pelo agente ao WhatsApp devem ser prefixadas com `*NomeDoAgente:*` seguido do delimitador configurado.

#### Scenario: texto outbound com signMsg ativo
- **WHEN** `cfg.SignMsg == true`, `senderName != ""`, e o conteúdo é texto
- **THEN** o texto enviado ao WA é `*{senderName}:*{delimiter}{content}`, onde `delimiter` é `cfg.SignDelimiter` (com `\n` literal transformado em newline real) ou `\n` se vazio

#### Scenario: attachment outbound com signMsg ativo
- **WHEN** `cfg.SignMsg == true`, `senderName != ""`, e a mensagem tem attachment
- **THEN** o caption do attachment é prefixado com `*{senderName}:*{delimiter}` antes de ser enviado ao WA

#### Scenario: signMsg com senderName vazio
- **WHEN** `cfg.SignMsg == true` mas `body.Conversation.Messages[0].Sender.AvailableName` e `body.Sender.Name` são ambos vazios
- **THEN** nenhuma assinatura é adicionada; mensagem enviada sem modificação

#### Scenario: signMsg desabilitado
- **WHEN** `cfg.SignMsg == false`
- **THEN** nenhuma assinatura é adicionada; comportamento atual mantido

### Requirement: signDelimiter suporta \\n literal
- **WHEN** `cfg.SignDelimiter` contém a string `\n` (dois caracteres: barra invertida + n)
- **THEN** é transformado em newline real `\n` (um caractere) antes de ser usado como delimitador
