# Sistema de Temperatura por CEP com OpenTelemetry e Zipkin

Este projeto implementa um sistema distribuído em Go que recebe um CEP, identifica a cidade e retorna o clima atual (temperatura em Celsius, Fahrenheit e Kelvin) juntamente com a cidade. O sistema é composto por dois serviços e implementa tracing distribuído usando OpenTelemetry e Zipkin.

## Arquitetura

- **Serviço A**: Responsável pelo input e validação do CEP
- **Serviço B**: Responsável pela orquestração, busca de localização e temperatura
- **OpenTelemetry Collector**: Coleta e processa traces
- **Zipkin**: Interface para visualização dos traces

## Estrutura do Projeto

```
├── docker-compose.yml
├── otel-collector-config.yaml
├── service-a/
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
└── service-b/
    ├── Dockerfile
    ├── go.mod
    └── main.go
```

## Funcionalidades

### Serviço A (Porta 8080)
- Recebe requisições POST com CEP no formato: `{"cep": "12345678"}`
- Valida se o CEP tem 8 dígitos e é uma string
- Encaminha requisições válidas para o Serviço B
- Retorna erro 422 para CEPs inválidos

### Serviço B (Porta 8081)
- Recebe CEPs válidos do Serviço A
- Busca dados de localização via API ViaCEP
- Busca dados de temperatura via WeatherAPI
- Converte temperaturas para Celsius, Fahrenheit e Kelvin
- Retorna dados estruturados com cidade e temperaturas

## APIs Utilizadas

- **ViaCEP**: Para busca de dados de localização por CEP
- **WeatherAPI**: Para busca de dados de temperatura (opcional - usa dados simulados se não configurado)

## Observabilidade

O sistema implementa tracing distribuído com:
- Spans para medir tempo de resposta
- Propagação de contexto entre serviços
- Coleta via OpenTelemetry
- Visualização via Zipkin

## Como Executar

### Pré-requisitos

- Docker
- Docker Compose

### Configuração (Opcional)

1. Crie um arquivo `.env` baseado no `.env.example`:
```bash
cp .env.example .env
```

2. Configure sua chave da WeatherAPI (opcional):
```bash
WEATHER_API_KEY=sua_chave_aqui
```

### Executando o Sistema

1. Clone o repositório e navegue até a pasta:
```bash
cd goexpert-get-clima
```

2. Execute o sistema com Docker Compose:
```bash
docker-compose up --build
```

3. O sistema estará disponível em:
   - Serviço A: http://localhost:8080
   - Serviço B: http://localhost:8081
   - Zipkin UI: http://localhost:9411

## Testando o Sistema

### Teste Básico

```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"cep": "01310100"}'
```

**Resposta esperada (200):**
```json
{
  "city": "São Paulo",
  "temp_C": 25.0,
  "temp_F": 77.0,
  "temp_K": 298.0
}
```

### Teste CEP Inválido

```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"cep": "123"}'
```

**Resposta esperada (422):**
```json
{
  "message": "invalid zipcode"
}
```

### Teste CEP Não Encontrado

```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"cep": "00000000"}'
```

**Resposta esperada (404):**
```json
{
  "message": "can not find zipcode"
}
```

## Visualizando Traces

1. Acesse a interface do Zipkin: http://localhost:9411
2. Execute algumas requisições
3. Clique em "Run Query" para ver os traces
4. Clique em um trace para ver detalhes dos spans

## Códigos de Resposta

- **200**: Sucesso - retorna dados de cidade e temperatura
- **422**: CEP inválido (formato incorreto)
- **404**: CEP não encontrado
- **400**: Corpo da requisição inválido
- **405**: Método não permitido
- **500**: Erro interno do servidor

## Fórmulas de Conversão

- **Celsius para Fahrenheit**: F = C × 1.8 + 32
- **Celsius para Kelvin**: K = C + 273

## Desenvolvimento

### Executando Localmente

Para desenvolvimento local, você pode executar os serviços separadamente:

1. **Serviço A**:
```bash
cd service-a
go mod tidy
go run main.go
```

2. **Serviço B**:
```bash
cd service-b
go mod tidy
go run main.go
```

Certifique-se de configurar as variáveis de ambiente apropriadas.

## Dependências Principais

- Go 1.21
- OpenTelemetry Go SDK
- Zipkin
- Docker & Docker Compose

## Troubleshooting

### Problema: Serviços não conseguem se comunicar
- Verifique se todos os containers estão rodando: `docker-compose ps`
- Verifique os logs: `docker-compose logs service-a` ou `docker-compose logs service-b`

### Problema: Traces não aparecem no Zipkin
- Verifique se o OpenTelemetry Collector está rodando
- Verifique os logs do collector: `docker-compose logs otel-collector`
- Aguarde alguns segundos após fazer requisições

### Problema: Erro na API de clima
- Se você não configurou WEATHER_API_KEY, o sistema usa dados simulados
- Para dados reais, registre-se em https://www.weatherapi.com/ e configure a chave
