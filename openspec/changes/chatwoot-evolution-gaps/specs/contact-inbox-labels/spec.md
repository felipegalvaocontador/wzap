## ADDED Requirements

### Requirement: label com nome do inbox adicionada ao contato após criação
Quando um novo contato é criado no Chatwoot e `cfg.DatabaseURI` está configurado, o sistema deve adicionar uma label com o nome do inbox ao contato via SQL direto no banco do Chatwoot.

#### Scenario: contato criado com DatabaseURI configurado
- **WHEN** um novo contato é criado com sucesso e `cfg.DatabaseURI != ""`
- **THEN** a função `addLabelToContact` é chamada com `cfg.InboxName` e o ID do novo contato
- **THEN** as tabelas `tags` e `taggings` do banco Chatwoot são atualizadas: `tags.name = inboxName`, `taggings.taggable_type = 'Contact'`, `taggings.context = 'labels'`

#### Scenario: label já existe para o contato
- **WHEN** a relação `taggings` para o contato e a tag já existe
- **THEN** nenhum INSERT duplicado é feito; a função retorna sem erro

#### Scenario: DatabaseURI não configurado
- **WHEN** `cfg.DatabaseURI == ""`
- **THEN** `addLabelToContact` não é chamada; nenhum erro é gerado; o fluxo prossegue normalmente

#### Scenario: erro ao adicionar label
- **WHEN** a conexão com o banco falha ou as queries retornam erro
- **THEN** o erro é logado como warn; a criação do contato e da conversa **não** são afetadas
