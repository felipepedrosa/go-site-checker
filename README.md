# 🔍 URL Status Checker

> 📚 **Projeto de estudos em Go** — desenvolvido por um dev Java/Kotlin explorando o ecossistema Golang.

---

## Sobre o projeto

Este projeto é um **exercício prático de aprendizado em Go**, criado com o objetivo de explorar conceitos fundamentais da linguagem que diferem do mundo Java/Kotlin, como goroutines, WaitGroups, tratamento de erros sem exceções e o uso da stdlib.

A aplicação recebe uma lista de URLs via linha de comando e verifica o **status HTTP** de cada uma delas de forma **concorrente**, exibindo os resultados em uma tabela formatada no terminal.

---

## O que ele faz

- Valida se as URLs fornecidas são válidas (via regex)
- Realiza requisições HTTP GET com timeout de 2 segundos
- Executa as verificações em **paralelo** usando goroutines (`sync.WaitGroup`)
- Filtra resultados inválidos ou com erro
- Exibe uma tabela formatada com endereço e status HTTP

### Exemplo de saída

```
Address                  | Status
-------------------------+--------
https://google.com       | 200
https://github.com       | 200
https://naoexiste.xyz    | (erro no stderr)
```

---

## Como executar

**Pré-requisito:** ter o [Go](https://go.dev/dl/) instalado (versão 1.22+).

```bash
# Clonar o repositório
git clone <seu-repo>
cd <seu-repo>

# Rodar diretamente
go run main.go https://google.com https://github.com https://example.com

# Ou compilar e executar
go build -o url-checker
./url-checker https://google.com https://github.com
```

---

## Conceitos de Go explorados

Como dev Java/Kotlin, estes foram os pontos de maior atenção e aprendizado:

| Conceito Go | Equivalente Java/Kotlin |
|---|---|
| `goroutines` + `sync.WaitGroup` | `CompletableFuture` / `coroutines` |
| Múltiplos retornos `(T, error)` | `throws` / `Result<T>` / exceções |
| `defer` para fechar recursos | `try-with-resources` |
| `make([]T, n)` | `new ArrayList<>()` / `Array(n)` |
| `os.Args` | `args[]` no `main` |
| `fmt.Fprintf(os.Stderr, ...)` | `System.err.println(...)` |
| Struct sem herança | POJO / data class |



## Próximos passos / ideias

- [x] Adicionar flag `-timeout` para configurar o timeout via CLI
- [x] Suportar leitura de URLs a partir de uma listagem
- [ ] Suportar leitura de URLs a partir de um arquivo
- [ ] Escrever testes unitários com o pacote `testing`
- [ ] Exportar resultado em JSON ou CSV

