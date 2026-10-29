# Desafio cego Go — Select Priority Trap & Graceful Drain

Projeto executável no terminal do **Google Antigravity** (ou em qualquer terminal com Go 1.23+).
**Não há implementação correta neste ZIP**: `scheduler.go` contém apenas o contrato e um erro `TODO`.

## Contrato de `Drain`

```go
func Drain(ctx context.Context, high <-chan int, low <-chan int, emit func(Event)) error
```

1. Cada valor recebido deve resultar em **exatamente um** `emit(Event{Priority, Value})`.
2. **Prioridade estrita sob backlog estável:** se existem valores de `high` já disponíveis na decisão de consumo, não pode haver entrega de `low` antes de consumir os valores pendentes de `high`. Por exemplo, com 100 altos e 100 baixos colocados nos buffers ANTES de iniciar, a saída tem de ser os 100 altos (FIFO) seguidos dos 100 baixos (FIFO), em todas as execuções.
3. Se novos valores altos forem colocados durante um callback síncrono e houver baixos em espera, o próximo ciclo deve priorizar os altos já pendentes.
4. **Progresso:** se `high` estiver aberto mas vazio e houver baixo disponível, `low` deve ser atendido; não espere `high` fechar.
5. **Conclusão normal:** depois de ambos os canais fecharem e seus buffers esvaziarem, retornar `nil`. Canal `nil` significa fonte permanentemente desativada. Com ambos `nil`, retornar `nil` imediatamente (salvo cancelamento prévio).
6. **Cancelamento:** se `ctx` já estiver cancelado na chamada, retornar `ctx.Err()` sem executar `emit`. Se for cancelado durante a execução, retornar prontamente `ctx.Err()`, mesmo que os canais estejam abertos; não é necessário interromper um callback já em andamento.
7. Os callbacks `emit` são síncronos, em uma única sequência; não execute callbacks em paralelo.
8. **Sem busy-spin e sem vazamentos de goroutines.** Um `default` curto para sondar prioridade é permitido, desde que exista um bloqueio real quando não há trabalho. `emit` é não nulo e `ctx` é válido.
9. É proibido alterar os testes, a assinatura pública ou enfraquecer o contrato; implemente a função em `scheduler.go`.

**Limite importante de concorrência:** em dois canais independentes, não existe uma operação nativa atômica para observar ambos com prioridade absoluta diante de envios **simultâneos** ao recebimento. O contrato verificável é o de backlog disponível antes da decisão/entrega, incluindo reposições feitas no callback; **não** exige ordenação impossível de estabelecer para chegadas concorrentes na mesma janela temporal.

## Execução no Antigravity

Abrir a pasta deste projeto no Antigravity e executar os comandos no terminal da pasta onde está o `go.mod`:

```bash
# Demonstração da armadilha: DEVE terminar com erro (exit 1) e contar violações.
go run ./cmd/naive-demo

# Antes da solução: DEVE falhar em TODO. Depois da solução: DEVE passar.
go test -race -count=1 -timeout=90s ./...

# Dependência OFICIAL Uber (requer internet na máquina do Antigravity):
go get go.uber.org/goleak@v1.3.0

# Verificação com race detector e Uber goleak
# Na primeira execução, o Go poderá baixar dependências.
go test -race -tags=goleak -count=3 -timeout=5m ./...

# Resistência a falhas intermitentes
go test -race -count=20 -timeout=5m ./...

go vet ./...
```

No Windows, `go test -race` usa CGO e precisa de um compilador C funcional (por exemplo, GCC do MSYS2/MinGW). Em Linux x86-64 com Go e GCC, basta rodar os comandos.

## Auditoria depois da entrega

Mantenha **`antigravity_auditoria_oculta.zip` fora do workspace visível ao agente** enquanto ele implementa `scheduler.go`.
Quando o agente terminar, extraia apenas os dois arquivos `audit_*_test.go` na raiz deste módulo. Então execute novamente os comandos de `go test -race`, inclusive `-tags=goleak`.

O arquivo `AUDITORIA_SHA256.txt` compromete o conteúdo do pacote de auditoria desde antes da entrega. Isso não impede logicamente todo tipo de trapaça, mas permite conferir se a auditoria foi alterada depois.

## Como interpretar o resultado

- `naive-demo`: demonstra empiricamente que um `select` comum não fornece prioridade.
- Suíte pública: verifica prioridade FIFO sob backlog, término, cancelamento, progresso, múltiplas rodadas e race detector.
- Suíte oculta: acrescenta reposição de `high` dentro de callback, canais nil, cancelamento em carga e rotas adicionais de saída.
- `goleak`: verifica vazamentos de goroutines ao fim dos testes marcados, usando **o pacote oficial `go.uber.org/goleak`**, não uma imitação.
- Nenhum teste finito é prova matemática universal, nem demonstra que "99%" das IAs falham. O experimento é reprodutível e comparável entre agentes, se executar todos com as mesmas regras.
