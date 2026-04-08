## ADDED Requirements

### Requirement: contactMessage formatado com markdown
Quando uma mensagem de contato (`contactMessage`) é recebida, o texto exibido no Chatwoot deve mostrar o nome do contato e todos os telefones numerados, usando markdown bold/italic do Chatwoot.

#### Scenario: contactMessage com nome e múltiplos telefones
- **WHEN** `contactMessage` com vCard contendo `FN` e múltiplos campos `TEL` é recebido
- **THEN** o texto enviado ao Chatwoot é:
  ```
  *Contato:*

  _Nome:_ {FN do vCard}
  _Número (1):_ {tel1}
  _Número (2):_ {tel2}
  ```

#### Scenario: contactMessage sem telefone
- **WHEN** `contactMessage` com vCard contendo apenas `FN`, sem `TEL`
- **THEN** apenas o cabeçalho e nome são exibidos, sem linhas de número

### Requirement: contactsArrayMessage formata múltiplos contatos
Quando uma lista de contatos (`contactsArrayMessage`) é recebida, cada contato é formatado individualmente e exibido em sequência separada por linha em branco dupla.

#### Scenario: contactsArrayMessage com dois contatos
- **WHEN** `contactsArrayMessage` com `contacts[0]` e `contacts[1]` é recebido
- **THEN** o texto resulta em dois blocos `*Contato:*` separados por `\n\n`, cada um com nome (de `displayName` com fallback para `FN` do vCard) e telefones numerados

#### Scenario: nome do contato — prioridade displayName
- **WHEN** o objeto `contacts[i]` tem `displayName` e o vCard tem `FN` diferentes
- **THEN** `displayName` é usado como nome (não `FN`)
