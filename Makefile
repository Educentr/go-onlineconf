PWD = $(CURDIR)
# Таймаут для тестов
TEST_TIMEOUT?=30s
# Тег golang-ci
GOLANGCI_TAG:=1.59.1
# Путь до бинарников
LOCAL_BIN:=$(CURDIR)/bin
# Путь до бинарника golang-ci
GOLANGCI_BIN:=$(LOCAL_BIN)/golangci-lint
# Минимальная версия Go
MIN_GO_VERSION = 1.21.5

# Добавляет флаг для тестирования на наличие гонок
ifdef GO_RACE_DETECTOR
    FLAGS += -race
endif

##################### Проверки для запуска golang-ci #####################
# Проверка локальной версии бинаря
ifneq ($(wildcard $(GOLANGCI_BIN)),)
GOLANGCI_BIN_VERSION:=$(shell $(GOLANGCI_BIN) --version)
ifneq ($(GOLANGCI_BIN_VERSION),)
GOLANGCI_BIN_VERSION_SHORT:=$(shell echo "$(GOLANGCI_BIN_VERSION)" | sed -E 's/.* version (.*) built from .* on .*/\1/g')
else
GOLANGCI_BIN_VERSION_SHORT:=0
endif
ifneq "$(GOLANGCI_TAG)" "$(word 1, $(sort $(GOLANGCI_TAG) $(GOLANGCI_BIN_VERSION_SHORT)))"
GOLANGCI_BIN:=
endif
endif

# Проверка глобальной версии бинаря
ifneq (, $(shell which golangci-lint))
GOLANGCI_VERSION:=$(shell golangci-lint --version 2> /dev/null )
ifneq ($(GOLANGCI_VERSION),)
GOLANGCI_VERSION_SHORT:=$(shell echo "$(GOLANGCI_VERSION)"|sed -E 's/.* version (.*) built from .* on .*/\1/g')
else
GOLANGCI_VERSION_SHORT:=0
endif
ifeq "$(GOLANGCI_TAG)" "$(word 1, $(sort $(GOLANGCI_TAG) $(GOLANGCI_VERSION_SHORT)))"
GOLANGCI_BIN:=$(shell which golangci-lint)
endif
endif
##################### Конец проверок golang-ci #####################

# Устанавливает линтер
.PHONY: install-lint
install-lint:
ifeq ($(wildcard $(GOLANGCI_BIN)),)
	$(info #Downloading golangci-lint v$(GOLANGCI_TAG))
	tmp=$$(mktemp -d) && cd $$tmp && pwd && go mod init temp && go get -d github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_TAG) && \
		go build -ldflags "-X 'main.version=$(GOLANGCI_TAG)' -X 'main.commit=test' -X 'main.date=test'" -o $(LOCAL_BIN)/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint && \
		rm -rf $$tmp
GOLANGCI_BIN:=$(LOCAL_BIN)/golangci-lint
endif

# Линтер проверяет лишь отличия от main
.PHONY: lint
lint: install-lint
	$(GOLANGCI_BIN) run --config=.golangci.yaml ./... --new-from-rev=origin/main

# Линтер проходится по всему коду
.PHONY: full-lint
full-lint: install-lint
	$(GOLANGCI_BIN) run --config=.golangci.yaml ./...

# Линтер с автоисправлением
.PHONY: lint-fix
lint-fix: install-lint
	$(GOLANGCI_BIN) run --fix --config=.golangci.yaml ./...

# Запуск unit тестов
.PHONY: test
test:
	./scripts/goversioncheck.sh $(MIN_GO_VERSION)
	go test $(FLAGS) -parallel=10 $(PWD)/... -coverprofile=cover.out -timeout=$(TEST_TIMEOUT)

# Тесты с детектором гонок
.PHONY: race
race:
	go test ./... -v -race -parallel=10

# Создание отчета о покрытии тестами
.PHONY: cover
cover:
	go test -timeout=$(TEST_TIMEOUT) -v -coverprofile=cover.out ./... && go tool cover -html=cover.out

# Текстовый отчёт о покрытии
.PHONY: txt-coverage
txt-coverage:
	go test -timeout=$(TEST_TIMEOUT) -coverprofile=cover.out ./...
	go tool cover -func cover.out | grep "total:"

# Установка mockgen
bin/: ; mkdir -p $@
bin/mockgen: | bin/
	GOBIN="$(realpath $(dir $@))" go install github.com/golang/mock/mockgen@v1.6.0

# Устанавливает pre-commit хук (запускает линтер)
.PHONY: pre-commit-hook
pre-commit-hook:
	touch ./.git/hooks/pre-commit
	echo '#!/bin/sh' > ./.git/hooks/pre-commit
	echo 'make lint' >> ./.git/hooks/pre-commit
	chmod +x ./.git/hooks/pre-commit

# Устанавливает pre-push хук (запускает тесты с покрытием)
.PHONY: pre-push-hook
pre-push-hook:
	touch ./.git/hooks/pre-push
	echo '#!/bin/sh' > ./.git/hooks/pre-push
	echo 'make cover' >> ./.git/hooks/pre-push
	chmod +x ./.git/hooks/pre-push

# Запуск MkDocs сервера для документации
.PHONY: run-mkdoc
run-mkdoc:
	docker run --rm -p 8000:8000 -v $(PWD):/docs squidfunk/mkdocs-material serve --dev-addr=0.0.0.0:8000 --watch-theme

# Сборка документации
.PHONY: build-docs
build-docs:
	docker run --rm -v $(PWD):/docs squidfunk/mkdocs-material build --strict
