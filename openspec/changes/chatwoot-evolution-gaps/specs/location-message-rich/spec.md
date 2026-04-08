## ADDED Requirements

### Requirement: locationMessage formatado como bloco estruturado
Quando uma mensagem de localização (`locationMessage`) é recebida do WhatsApp, o texto exibido no Chatwoot deve conter latitude, longitude, nome do local (se presente), endereço (se presente) e um link Google Maps clicável, usando markdown Chatwoot (bold/italic).

#### Scenario: locationMessage com todos os campos
- **WHEN** `locationMessage` com `degreesLatitude`, `degreesLongitude`, `name` e `address` é recebido
- **THEN** o texto enviado ao Chatwoot é:
  ```
  *Localização:*

  _Latitude:_ {lat}
  _Longitude:_ {lng}
  _Nome:_ {name}
  _Endereço:_ {address}
  _URL:_ https://www.google.com/maps/search/?api=1&query={lat},{lng}
  ```

#### Scenario: locationMessage sem nome e endereço
- **WHEN** `locationMessage` com apenas `degreesLatitude` e `degreesLongitude`
- **THEN** linhas `_Nome:_` e `_Endereço:_` são omitidas; demais campos presentes

#### Scenario: liveLocationMessage
- **WHEN** `liveLocationMessage` com qualquer combinação de campos é recebido
- **THEN** aplica o mesmo template do `locationMessage`; sem mensagem adicional sobre atualizações subsequentes

### Requirement: URL Google Maps usa endpoint de busca
- **WHEN** o link Google Maps é construído
- **THEN** a URL base é `https://www.google.com/maps/search/?api=1&query=` com os parâmetros `{lat},{lng}`
