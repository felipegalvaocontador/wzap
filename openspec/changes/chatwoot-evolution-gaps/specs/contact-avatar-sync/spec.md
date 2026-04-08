## ADDED Requirements

### Requirement: avatar_url como campo de primeiro nível em CreateContact
Quando um contato é criado no Chatwoot, o campo `avatar_url` deve ser enviado diretamente no root do payload JSON (não dentro de `additional_attributes`).

#### Scenario: criar contato com foto de perfil disponível
- **WHEN** a foto de perfil WA do contato é obtida com sucesso antes de criar o contato
- **THEN** `CreateContactReq.AvatarURL` é preenchido com a URL da foto e enviado ao Chatwoot

#### Scenario: criar contato sem foto de perfil
- **WHEN** a obtenção de foto de perfil WA falha ou retorna vazio
- **THEN** o contato é criado normalmente sem `avatar_url`; o erro de obtenção de foto não bloqueia a criação

### Requirement: avatar_url como campo de primeiro nível em UpdateContact
Quando o avatar de um contato é atualizado (evento `EventPicture`), o campo `avatar_url` deve ser enviado diretamente no root do payload, não em `additional_attributes`.

#### Scenario: evento EventPicture recebido
- **WHEN** evento de atualização de foto de perfil WA chega
- **THEN** `UpdateContactReq.AvatarURL` é preenchido com a nova URL e o campo `additional_attributes.avatar_url` **não** é enviado

### Requirement: busca de foto de perfil no momento de criação de contato
- **WHEN** o slow path de criação de conversa é executado e um contato novo precisa ser criado
- **THEN** a foto de perfil WA é buscada antes da chamada `CreateContact`; timeout de no máximo 5s para a busca
