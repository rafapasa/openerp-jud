# ================================================================
# Makefile - MCP Server openerp-jud (eProc TJSC / Projudi TJPR)
# ================================================================
export PATH := $(HOME)/go/bin:$(PATH)
DOCKER_USERNAME := rafapasa
IMAGE_TAG ?= latest
PLATFORM := linux/arm64
DOCKERFILE := Dockerfile.multistage
DOCKERFILE_TESTE := Dockerfile.teste.multistage
NO_CACHE := --no-cache

APP_NAME := openerp-jud
IMAGE_PROD := $(DOCKER_USERNAME)/$(APP_NAME)
IMAGE_TEST := $(DOCKER_USERNAME)/$(APP_NAME)-test

# Paths de deploy (espelha o padrão openerp-whatsapp)
DEPLOY_DIR_PROD := /home/opc/etools/openerp-jud
DEPLOY_DIR_TEST := /home/opc/prj/openerp-jud

# ================================================================
# DESENVOLVIMENTO LOCAL
# ================================================================

.PHONY: build
build: wire test fmt vet build-bin

.PHONY: build-bin
build-bin:
	go build -o bin/$(APP_NAME) .

.PHONY: run
run:
	go run .

.PHONY: run-sse
run-sse:
	MCP_TRANSPORT=sse PORT=8080 go run .

.PHONY: run-stdio
run-stdio:
	MCP_TRANSPORT=stdio go run .

.PHONY: dev wire
dev:
	air

wire:
	rm -f internal/di/wire_gen.go && wire gen ./internal/di/...

.PHONY: test fmt vet lint
test:
	go test -v -coverprofile=coverage.out ./...
	@echo ""
	@echo "========== COBERTURA =========="
	@go tool cover -func=coverage.out | tail -1
	@echo "==============================="

fmt:
	gofmt -w -s .
	goimports -w .

vet:
	go vet ./...

lint:
	golangci-lint run ./...

.PHONY: deps install-tools
deps:
	go mod download && go mod tidy

install-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/air-verse/air@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: playwright-install
playwright-install:
	go run github.com/mxschmitt/playwright-go/cmd/playwright@v0.6401.0 install chromium

.PHONY: clean
clean:
	rm -rf bin/ tmp/ coverage.out
	go clean

# ================================================================
# 🚀 DEPLOY OCI - ORACLE LINUX 9 / AMPERE ARM64
# ================================================================

.PHONY: login
login:
	docker login -u $(DOCKER_USERNAME)

.PHONY: init-network
init-network:
	@echo "🐳 Criando rede mcp-network..."
	docker network create mcp-network || true
	@echo "✅ Rede mcp-network pronta"

# Build prod — exige IMAGE_TAG versionado (nunca só latest)
.PHONY: build-push
build-push:
ifeq ($(IMAGE_TAG),latest)
	$(error Use: make build-push IMAGE_TAG=0.1.0 - não pode buildar só latest)
endif
	git pull
	DOCKER_BUILDKIT=1 docker build $(NO_CACHE) -f $(DOCKERFILE) --target server \
		-t $(IMAGE_PROD):$(IMAGE_TAG) \
		-t $(IMAGE_PROD):latest .
	docker push $(IMAGE_PROD):$(IMAGE_TAG)
	docker push $(IMAGE_PROD):latest
	@echo "✅ Build $(IMAGE_TAG) + latest enviado para $(IMAGE_PROD)"

.PHONY: deploy
deploy:
	@echo "🚀 Deploy $(APP_NAME) IMAGE_TAG=$(IMAGE_TAG)..."
	@cd $(DEPLOY_DIR_PROD) && \
		IMAGE_TAG=$(IMAGE_TAG) docker compose --env-file $(DEPLOY_DIR_PROD)/.env -f docker-compose.app.yml up -d --pull always --no-deps $(APP_NAME)
	@docker ps | grep $(APP_NAME) || true

# Build/teste — tag fixa "test"
.PHONY: build-push-t
build-push-t:
	DOCKER_BUILDKIT=1 docker build $(NO_CACHE) -f $(DOCKERFILE_TESTE) --target server \
		-t $(IMAGE_TEST):test .
	docker push $(IMAGE_TEST):test
	@echo "✅ Build test enviado para $(IMAGE_TEST):test"

.PHONY: deploy-t
deploy-t:
	@echo "🚀 Deploy $(APP_NAME)-test..."
	@cd $(DEPLOY_DIR_TEST) && \
		IMAGE_TAG=test docker compose --env-file $(DEPLOY_DIR_TEST)/.env -f docker-compose.app.teste.yml up -d --pull always --no-deps $(APP_NAME)-test
	@docker ps | grep $(APP_NAME)-test || true

.PHONY: teste
teste: build-push-t deploy-t

# ================================================================
# LOGS / STATUS
# ================================================================

.PHONY: logs logs-t logs-tail logs-info logs-warn logs-err logs-debug ps status

logs:
	docker logs -f $(APP_NAME) --tail=100

logs-t:
	docker logs -f $(APP_NAME)-test --tail=100

logs-tail:
	docker logs $(APP_NAME) --tail=200

logs-info:
	docker logs -f $(APP_NAME) --tail=500 2>&1 | grep --color=always -E "INFO|Iniciando|MCP|browser"

logs-warn:
	docker logs -f $(APP_NAME) --tail=500 2>&1 | grep --color=always -E "WARN"

logs-err:
	docker logs -f $(APP_NAME) --tail=500 2>&1 | grep --color=always -E "ERROR|ERRO|FAIL|✖|panic|fatal"

logs-debug:
	docker logs -f $(APP_NAME) --tail=500 2>&1 | grep --color=always -E "DEBUG"

ps status:
	docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "mcp|openerp-jud" || true

# ================================================================
# GIT
# ================================================================

.PHONY: git-status git-add git-commit git-push git-push-main git-merge-origin-main git-restore-wire git-commit-files

git-status:
	git status -sb && git diff --stat

git-add:
	git add -A && git status -sb

git-commit:
ifndef MSG
	$(error Use: make git-commit MSG='tipo: mensagem')
endif
	git add -A
	git commit -m "$(MSG)"

git-push:
	git push -u origin HEAD

git-push-main:
	git push origin HEAD:main

git-merge-origin-main:
	git fetch origin main
	git merge --no-edit origin/main

git-restore-wire:
	git checkout -- internal/di/wire_gen.go

git-commit-files:
ifndef MSG
	$(error Use: make git-commit-files MSG='tipo: mensagem' FILES='a.go b.go')
endif
	git add $(FILES)
	git commit -m "$(MSG)"
