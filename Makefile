.PHONY: build up down logs test clean deps

# Build e execução
build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

# Teste do sistema
test:
	./test.sh

# Limpeza
clean:
	docker-compose down -v
	docker system prune -f

# Atualizar dependências
deps:
	cd service-a && go mod tidy
	cd service-b && go mod tidy

# Executar tudo
run: build up

# Verificar status dos serviços
status:
	docker-compose ps

# Abrir Zipkin no navegador (macOS)
zipkin:
	open http://localhost:9411

# Rebuild completo
rebuild: down clean build up
