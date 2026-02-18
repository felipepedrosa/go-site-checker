# 🔍 URL Status Checker

> 📚 **Projeto de estudos em Go** — desenvolvido por um dev Java/Kotlin explorando o ecossistema Golang.

---

## Sobre o projeto

Este projeto é um **exercício prático de aprendizado em Go**, criado com o objetivo de explorar conceitos fundamentais da linguagem que diferem do mundo Java/Kotlin, como goroutines, WaitGroups, tratamento de erros sem exceções e o uso da stdlib.

A aplicação recebe uma lista de URLs via linha de comando ou arquivo JSON e verifica o **status HTTP** de cada uma delas de forma **concorrente**, exibindo os resultados em uma tabela formatada no terminal ou exportando para CSV.

---

## O que ele faz

- Valida se as URLs fornecidas são válidas (via `url.ParseRequestURI`)
- Realiza requisições HTTP GET com timeout configurável (padrão: 2 segundos)
- Executa as verificações em **paralelo** usando goroutines e `sync.WaitGroup`
- Filtra resultados inválidos ou com erro
- Suporta múltiplas fontes de entrada: linha de comando ou arquivo JSON
- Exibe uma tabela formatada com endereço e status HTTP
- Exporta resultados para CSV quando solicitado

### Exemplo de saída (modo tabela)
```
URL                      | Status
-------------------------+--------
https://google.com       | 200
https://github.com       | 200
-------------------------+--------
```

### Exemplo de saída (modo export)
```
results exported to output.csv
```

**Conteúdo do CSV:**
```csv
URL;Status
https://google.com;200
https://github.com;200
```

---

## Como executar

**Pré-requisito:** ter o [Go](https://go.dev/dl/) instalado (versão 1.22+).
```bash
# Clonar o repositório
git clone <seu-repo>
cd <seu-repo>
```

### Modo 1: URLs via linha de comando
```bash
# Rodar diretamente
go run main.go --addresses "https://google.com;https://github.com;https://example.com"

# Com timeout customizado (5 segundos)
go run main.go --addresses "https://google.com;https://github.com" --timeout 5

# Exportar para CSV
go run main.go --addresses "https://google.com;https://github.com" --export
```

### Modo 2: URLs via arquivo JSON

Crie um arquivo `addresses.json`:
```json
[
  "https://google.com",
  "https://github.com",
  "https://example.com"
]
```

Execute:
```bash
# Usar arquivo padrão (addresses.json)
go run main.go --file

# Usar arquivo customizado
go run main.go --file --filepath meus-enderecos.json

# Ler de arquivo e exportar para CSV
go run main.go --file --export
```

### Modo 3: Compilar e executar
```bash
go build -o url-checker
./url-checker --addresses "https://google.com;https://github.com"
```

---

## Flags disponíveis

| Flag | Tipo | Padrão | Descrição |
|------|------|--------|-----------|
| `--addresses` | string | - | URLs separadas por `;` (ex: `"url1;url2"`) |
| `--file` | bool | false | Ler URLs de um arquivo JSON |
| `--filepath` | string | `addresses.json` | Caminho do arquivo JSON (usar com `--file`) |
| `--timeout` | int | 2 | Timeout das requisições HTTP em segundos |
| `--export` | bool | false | Exportar resultados para CSV (`output.csv`) |

### Validações

- Não é possível usar `--addresses` e `--file` simultaneamente
- Pelo menos uma das flags `--addresses` ou `--file` deve ser fornecida
- O `--timeout` deve ser maior que 0
- Arquivo JSON deve ter extensão `.json`

---

## Conceitos de Go explorados

Como dev Java/Kotlin, estes foram os pontos de maior atenção e aprendizado:

| Conceito Go | Equivalente Java/Kotlin |
|---|---|
| `goroutines` + `sync.WaitGroup` | `CompletableFuture` / `coroutines` |
| `sync.Mutex` para controle de concorrência | `synchronized` / `Mutex` |
| Múltiplos retornos `(T, error)` | `throws` / `Result<T>` / exceções |
| `defer` para fechar recursos | `try-with-resources` / `use` |
| `make([]T, n)` e `append` | `new ArrayList<>()` / `mutableListOf()` |
| `flag` package | `Apache Commons CLI` / `kotlinx-cli` |
| `os.Stderr` vs `os.Stdout` | `System.err` vs `System.out` |
| `bufio.Writer` para I/O eficiente | `BufferedWriter` |
| Structs sem herança | POJO / data class |
| JSON marshaling/unmarshaling | `Jackson` / `Gson` / `kotlinx.serialization` |

---

## Estrutura do código
```
main.go
├── main()                      # Entry point e orquestração
├── configureFlags()            # Configuração e validação de flags
├── getAddresses()              # Obtenção de URLs (CLI ou arquivo)
├── checkAddresses()            # Verificação concorrente com goroutines
├── printStatuses()             # Exibição em tabela formatada
├── exportToCSV()               # Exportação para CSV
└── getUrlStatus()              # Request HTTP individual
```

---

## Próximos passos / ideias

- [x] Adicionar flag -timeout para configurar o timeout via CLI
- [x] Suportar leitura de URLs a partir de uma listagem
- [x] Suportar leitura de URLs a partir de um arquivo
- [ ] Escrever testes unitários com o pacote testing
- [x] Exportar resultado em JSON ou CSV
- [ ] Usar channels ao invés de slices para os endereços processados

---

## Aprendizados

### 1. Concorrência é simples em Go
Goroutines + WaitGroup tornaram trivial o processamento paralelo que em Java exigiria muito mais boilerplate.

### 2. Tratamento de erros explícito
Sem `try/catch`, cada função retorna `(resultado, error)`. Isso força o dev a pensar em cada ponto de falha.

### 3. Flags nativas
O package `flag` é simples mas poderoso. Em Java, precisaríamos de bibliotecas externas como Apache Commons CLI.

### 4. I/O é direto
`bufio.Writer`, `os.Stderr`, `defer file.Close()` — tudo muito explícito e sem "mágica".

---

## Licença

Projeto de estudos — use à vontade! 🚀