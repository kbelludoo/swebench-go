# Copie este texto na conversa com o agente Antigravity

Você está recebendo um desafio **cego** de engenharia Go, sem implementação pronta.
Leia `README.md`, `scheduler.go` e os testes públicos antes de programar.

OBJETIVO: implementar `Drain` em `scheduler.go` com prioridade estrita quando `high` já tem backlog, preservação FIFO por canal, saída por `context.Context`, fechamento gracioso, ausência de busy-spin e ausência de vazamento de goroutines. Canais `nil` são fontes desativadas. Os callbacks devem ser síncronos. Observe a ressalva do README sobre envios simultâneos e impossibilidade de ordem atômica global em canais independentes.

REGRAS:
- Escreva a solução do zero; nenhuma implementação correta está incluída neste workspace.
- NÃO modifique testes, assinaturas ou arquivos do auditor. Modifique **somente `scheduler.go`**, exceto a alteração automática em `go.mod`/`go.sum` necessária para instalar o `goleak` oficial.
- NÃO simule resultados, não comente/remova testes, não crie atalhos por contagem de mensagens, nem use `time.Sleep` para fabricar prioridade.
- Pode usar as construções normais da biblioteca padrão de Go. Não crie spin-loop com `default` sem espera bloqueante.
- Primeiro execute `go test -race -count=1 -timeout=90s ./...`, constate a falha TODO, implemente, teste de novo.
- Instale a dependência original com `go get go.uber.org/goleak@v1.3.0` e rode `go test -race -tags=goleak -count=3 -timeout=5m ./...`.
- Ao final, informe **comandos executados, logs efetivos, commit/diff de `scheduler.go`, quaisquer falhas remanescentes e limitações**. Não declare testes não executados como aprovados.
- Quando você disser que concluiu, vou inserir testes privados adicionais que você não pode ver antecipadamente.
