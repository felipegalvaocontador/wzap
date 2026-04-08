## ADDED Requirements

### Requirement: viewOnceMessageV2 envia o arquivo real
Quando uma mensagem de visualização única (`viewOnceMessageV2`) é recebida, o sistema deve tentar baixar e enviar o arquivo real (imagem, vídeo ou áudio) como anexo no Chatwoot, em vez de enviar apenas texto placeholder.

#### Scenario: viewOnce com imageMessage interno
- **WHEN** `viewOnceMessageV2.message.imageMessage` existe com `directPath`, `mediaKey` e `mimetype`
- **THEN** o arquivo é baixado via downloader existente e enviado como attachment no Chatwoot; o texto `[mensagem vista uma vez]` **não** é enviado

#### Scenario: viewOnce com videoMessage interno
- **WHEN** `viewOnceMessageV2.message.videoMessage` existe
- **THEN** comportamento idêntico ao scenario com imageMessage

#### Scenario: viewOnce com audioMessage interno
- **WHEN** `viewOnceMessageV2.message.audioMessage` existe
- **THEN** comportamento idêntico ao scenario com imageMessage

#### Scenario: viewOnce sem sub-mensagem reconhecível ou download falha
- **WHEN** `viewOnceMessageV2.message` é nil, não contém tipo reconhecível, ou o download falha por erro (chave expirada, timeout)
- **THEN** mensagem de texto `[mensagem vista uma vez]` é enviada como fallback; nenhum erro é propagado ao consumer NATS

### Requirement: viewOnceMessage (v1) mantém comportamento atual
- **WHEN** `viewOnceMessage` (sem V2) é recebido
- **THEN** comportamento atual mantido (texto placeholder); sem tentativa de download
