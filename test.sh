#!/bin/bash

# Script para testar o sistema de temperatura por CEP

echo "=== Teste do Sistema de Temperatura por CEP ==="
echo ""

# Aguardar serviços ficarem disponíveis
echo "Aguardando serviços ficarem disponíveis..."
sleep 10

# URL base do Serviço A
SERVICE_A_URL="http://localhost:8080"

echo "1. Testando CEP válido (01310-100 - São Paulo/SP):"
curl -X POST $SERVICE_A_URL \
  -H "Content-Type: application/json" \
  -d '{"cep": "01310100"}' \
  -w "\nStatus Code: %{http_code}\n\n"

echo "2. Testando CEP inválido (formato incorreto):"
curl -X POST $SERVICE_A_URL \
  -H "Content-Type: application/json" \
  -d '{"cep": "123"}' \
  -w "\nStatus Code: %{http_code}\n\n"

echo "3. Testando CEP não encontrado:"
curl -X POST $SERVICE_A_URL \
  -H "Content-Type: application/json" \
  -d '{"cep": "00000000"}' \
  -w "\nStatus Code: %{http_code}\n\n"

echo "4. Testando outro CEP válido (20040-020 - Rio de Janeiro/RJ):"
curl -X POST $SERVICE_A_URL \
  -H "Content-Type: application/json" \
  -d '{"cep": "20040020"}' \
  -w "\nStatus Code: %{http_code}\n\n"

echo "5. Testando CEP com letras (inválido):"
curl -X POST $SERVICE_A_URL \
  -H "Content-Type: application/json" \
  -d '{"cep": "0131010A"}' \
  -w "\nStatus Code: %{http_code}\n\n"

echo "=== Teste concluído ==="
echo ""
echo "Para visualizar os traces no Zipkin, acesse: http://localhost:9411"
