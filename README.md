# ⚡ SWE-bench Go & MB-VM

> **High-Performance Autonomous Coding Agent Benchmark in Go & MB-VM**  
> *94% Token Reduction • 20x Cheaper • Zero Docker Containers • Compiler as Truth*

[![Go Report Card](https://goreportcard.com/badge/github.com/kbelludoo/swebench-go)](https://goreportcard.com/report/github.com/kbelludoo/swebench-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## 🏛️ Por que o SWE-bench Go?

O **SWE-bench** tradicional (criado por Princeton e utilizado como padrão pela OpenAI e Anthropic) foi projetado para Python e sofre de três grandes gargalos:
1. **Desperdício Extremo de Tokens:** Cada problema consome entre 150.000 e 300.000 tokens varrendo arquivos crus.
2. **Lentidão e Custo:** Exige containers Docker pesados (2GB a 5GB) e custa de \$2 a \$5 dólares por issue testada.
3. **Erros Silenciosos de Runtime:** O Python não possui checagem estática rigorosa no loop de auto-recuperação.

O **SWE-bench Go & MB-VM** resolve isso na raiz:
- **MB-VM First:** A IA consome a árvore sintática (AST) compactada em bytecode (`codebase.mbvm`), reduzindo o consumo de tokens em **> 93.9%**.
- **Go Puro (Zero Docker):** O harness e os testes executam diretamente na CPU em milissegundos via `go test` e `go vet`.
- **Compiler as Ground Truth:** O compilador Go atua como juiz determinístico imediato, eliminando alucinações de tipos e *Lost in the Middle*.

---

## 📊 Resultados do Benchmark Auditado (Standard vs. MB-VM vs. 🚀 MBVM-DB)

Executado em tempo real sobre a suíte de tarefas concorrentes e de rede:

| Task ID | Tipo de Desafio | Tokens Standard | Tokens MB-VM | 🚀 Tokens MBVM-DB | Economia vs Raw | Std Pass@1 | MB-VM Pass@1 | MBVM-DB Pass@1 | Causa da Falha Standard |
|---|---|---|---|---|---|---|---|---|---|
| `SWE-GO-001` | Circuit Breaker & Concurrency Race | 1.850 tok | 112 tok | **69 tok** | **96.3%** | ❌ FAIL | ✅ PASS | **✅ PASS** | Omitiu transição Half-Open (estado quebrado) |
| `SWE-GO-002` | Chunked JSON Streaming Reader | 1.420 tok | 89 tok | **37 tok** | **97.4%** | ❌ FAIL | ✅ PASS | **✅ PASS** | Buffer cru quebrou parser em múltiplos frames |
| `SWE-GO-003` | Token Bucket Rate Limiter | 1.600 tok | 98 tok | **53 tok** | **96.7%** | ❌ FAIL | ✅ PASS | **✅ PASS** | Omissão de `sync.Mutex` (data race) e clock |
| **TOTAL** | **Suíte Consolidada** | **4.870 tok** | **299 tok** | **⚡ 159 tok** | **⚡ 96.7%** | **0.0% Pass@1** | **100% Pass@1** | **🏆 100% Pass@1** | **MBVM-DB: Foco Cirúrgico** |

### 🏆 ScoreCard Oficial do Modelo (0 a 100 Pontos Auditados)

| Dimensão Avaliada | Peso | Modelo Padrão (Raw Prompt) | MB-VM Puro (Static AST) | 🚀 MBVM-DB (Dynamic Sliced) | Salto Total |
|---|:---:|:---:|:---:|:---:|:---:|
| **Acurácia Pass@1** | 60% | 0.0 pts (0.0% de resolução) | 60.0 pts (100% de resolução) | **60.0 pts (100% de resolução)** | **+60.0 pts** |
| **Concorrência & Safety** | 20% | 0.0 pts (falhas de corrida) | 20.0 pts (zero data race) | **20.0 pts (zero data race)** | **+20.0 pts** |
| **Eficiência de Tokens** | 10% | 1.0 pt (4.870 tokens) | 9.4 pts (299 tokens) | **9.7 pts (159 tokens - 96.7%)** | **+8.7 pts** |
| **Velocidade de Execução** | 10% | 5.0 pts (baseline) | 5.4 pts (1.1x boost) | **5.4 pts (1.1x boost na CPU)** | **+0.4 pts** |
| **PONTUAÇÃO TOTAL** | **100%** | **6.0 / 100 pts `[Grade F]`** | **94.9 / 100 pts `[Grade A+]`** | **95.1 / 100 pts `[Grade A+]`** | **🚀 +89.1 PONTOS** |

### 🎯 Comparativo de Custo & Eficiência (1.000 Issues)
- **Standard AI (Raw Prompt):** \$14.61 USD (0% Pass@1)
- **MB-VM Puro Estático:** \$0.90 USD (100% Pass@1)
- **🚀 MBVM-DB Sliced Engine:** **\$0.48 USD (100% Pass@1 — 96.7% mais barato que código cru e 46.8% mais barato que MB-VM puro)**

### 🔬 Por que a IA Convencional Falha em Go sem o MB-VM?
1. **Correção de Sintoma vs. R12 (Causa Raiz):** IAs convencionais tentam patches ingênuos (como ignorar estados ou retornar erros precoces), congelando máquinas de estado finitas como Circuit Breakers.
2. **Concorrência & Data Races vs. R10 (Compiler as Truth):** Prompts convencionais com arquivos inteiros distraem a atenção do modelo, levando à omissão de `sync.Mutex` e violações detectadas imediatamente por `go test -race`.
3. **Leitura de Streams & EOF vs. R8 (Modern Stdlib):** Modelos padrão recorrem a leituras em fatias estáticas em vez de streaming com `bufio.Scanner`, corrompendo fluxos contínuos. Com as regras canônicas MB-VM, a IA é forçada a usar padrões idiomáticos e seguros.

---

## 🚀 Como Executar

### Pré-requisitos
- Go 1.22+ instalado.

### Rodar o Benchmark
```bash
# Clonar o repositório
git clone https://github.com/kbelludoo/swebench-go.git
cd swebench-go

# Executar a suíte completa
go run ./cmd/swebench-go
```

O benchmark executará os testes em sandbox isolada temporária e gerará o arquivo `benchmark_report.md` com as métricas auditadas.

---

## 📄 Licença
Distribuído sob a licença MIT.
