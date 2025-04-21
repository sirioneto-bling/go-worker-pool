[![Continuous Integration Status](https://github.com/sirioneto-bling/go-worker-pool/actions/workflows/ci.yml/badge.svg)](https://github.com/sirioneto-bling/go-worker-pool/actions/workflows/ci.yml)

# go-worker-pool

Este projeto implementa um **Worker Pool** em Go: uma técnica eficiente para gerenciar múltiplas goroutines executando tarefas concorrentes de forma controlada.

## 🎯 Objetivo

Fornecer uma estrutura reutilizável e robusta para:

- Executar tarefas em paralelo.
- Processar resultados com tratamento de erro.
- Encerramento controlado (graceful shutdown) do worker pool.
- Aplicar em cenários variados (requisições massivas via API, consumidores Kafka, etc).

## ✅ Funcionalidades

- **Concorrência segura** com goroutines e canais.
- **Workers configuráveis**: número de workers e tamanho do canal de tarefas.
- **Adição de tarefas com ou sem retorno de resultado**.
- **Tratamento de erros personalizado** por tarefa.
- **Canal de resultados** para processar saídas das tarefas.
- **Encerramento controlado**: aguarda o fim do processamento antes de liberar recursos.

## 🔧 Como usar

### 1. Criar o pool

```go
pool, err := work.NewPool(numWorkers, bufferSize)
```

### 2. Definir a tarefa

```go
job := work.NewJobWithResult(func() (any, error) {
    // lógica da tarefa
}, func(err error) {
    // tratamento de erro
})
```

### 3. Adicionar tarefas

```go
pool.Start(ctx)
for i := 0; i < 100; i++ {
    pool.AddJob(job)
}
```

### 4. Aguardar o fim e consumir resultados

```go
go func() {
    pool.WaitJobs()
    pool.Stop()
}()

for result := range pool.Result() {
    if result.Error != nil {
        // lidar com erro
    } else {
        // lidar com sucesso
    }
}
```

---

## 💡 Exemplo completo (main.go)

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sirioneto-bling/go-tour/work"
)

func main() {
	start := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := work.NewPool(100, 200)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.Default()
	job := work.NewJobWithResult(sendRequestWithResult, handleRequestError)

	pool.Start(ctx)

	// Envia 500 tarefas
	for i := 0; i < 500; i++ {
		pool.AddJob(job)
	}

	// Aguarda processamento completo e encerra o pool
	go func() {
		pool.WaitJobs()
		logger.Println("Finalizando pool após o processamento")
		pool.Stop()
	}()

	total, withError := 0, 0

	// Loop principal: aguarda conclusão ou cancelamento
	for {
		select {
		case <-ctx.Done():
			return
		case <-pool.Done():
			logger.Println("Total de tarefas processadas:", total)
			logger.Println("Com sucesso:", total-withError)
			logger.Println("Com erro:", withError)
			logger.Printf("Tempo de execução: %.2f segundos\n", time.Since(start).Seconds())
			return
		default:
			// Consome resultados
			for {
				result, ok := <-pool.Result()
				if !ok {
					break
				}

				total++
				if result.Error != nil {
					withError++
					logger.Println("[ERRO]", result.Error.Error())
				} else {
					logger.Println("[SUCESSO]", result.Value)
				}
			}
		}
	}
}

func sendRequestWithResult() (any, error) {
	const url = "https://www.google.com/"
	time.Sleep(2 * time.Second) // simula carga de trabalho
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%s retornou status %d", url, res.StatusCode), nil
}

func handleRequestError(err error) {
	log.Println("Erro durante requisição:", err)
}
```

---

## 🧪 Benchmark

Para rodar os testes de benchmark, execute:

```bash
go test -bench=. ./tests
```

---

## 🧵 Aplicações

- **APIs**: disparar múltiplos processos em paralelo em uma única requisição (ex: uploads, cálculos, etc).
- **Consumidores Kafka**: manter workers processando eventos recebidos continuamente.
- **ETL / Data pipelines**: paralelizar transformações e envio de dados.
