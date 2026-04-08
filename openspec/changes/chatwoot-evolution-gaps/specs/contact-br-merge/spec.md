## ADDED Requirements

### Requirement: merge de contatos BR via API Chatwoot
Quando exatamente duas variantes de número brasileiro (com e sem 9º dígito) são encontradas no Chatwoot para o mesmo contato, o sistema deve chamar a API de merge do Chatwoot para unificar os registros.

#### Scenario: dois contatos BR encontrados com e sem 9º dígito
- **WHEN** `cfg.MergeBRContacts == true`, o número começa com `55`, e exatamente 2 contatos são encontrados (um com 14 chars no phone_number, outro com 13)
- **THEN** `POST /api/v1/accounts/{accountId}/actions/contact_merge` é chamado com `base_contact_id` = ID do contato com 14 chars (com 9) e `mergee_contact_id` = ID do contato com 13 chars (sem 9)
- **THEN** após o merge, o contato com 9º dígito (base) é usado como `contactID` para a conversa

#### Scenario: merge falha com erro HTTP
- **WHEN** a chamada de merge retorna erro (422, 500, timeout)
- **THEN** o erro é logado como warn; o contato com 14 chars é usado como fallback sem interrupção do fluxo

#### Scenario: apenas um contato encontrado
- **WHEN** apenas uma variante do número é encontrada
- **THEN** nenhum merge é realizado; comportamento atual mantido

#### Scenario: MergeBRContacts desabilitado
- **WHEN** `cfg.MergeBRContacts == false`
- **THEN** nenhuma chamada de merge é feita; comportamento atual mantido
