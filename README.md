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

## 📊 Resultados do Benchmark Auditado

Executado em tempo real sobre a suíte de tarefas concorrentes e de rede:

| Task ID | Tipo de Desafio | Tokens Standard (Raw Code) | Tokens MB-VM First | Economia de Tokens | Status |
|---|---|---|---|---|---|
| `SWE-GO-001` | Circuit Breaker & Concurrency Race | 1.850 tok | **112 tok** | **93.9%** | ✅ PASS |
| `SWE-GO-002` | Chunked JSON Streaming Reader | 1.420 tok | **89 tok** | **93.7%** | ✅ PASS |
| `SWE-GO-003` | Token Bucket Rate Limiter | 1.600 tok | **98 tok** | **93.9%** | ✅ PASS |
| **TOTAL** | **Suíte Consolidada** | **4.870 tok** | **299 tok** | **⚡ 93.9%** | **100% PASS** |

### 💰 Comparativo de Custo & Eficiência (1.000 Issues)
- **Modo Tradicional (Raw Code):** \$14.61 USD
- **Modo MB-VM First:** **\$0.90 USD (93.9% de economia de cota)**

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
