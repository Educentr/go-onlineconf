PWD = $(CURDIR)
# Таймаут для тестов
TEST_TIMEOUT?=30s
# Тег golang-ci
GOLANGCI_TAG:=1.64.8
# Путь до бинарников
LOCAL_BIN:=$(CURDIR)/bin
# Путь до бинарника golang-ci
GOLANGCI_BIN:=$(LOCAL_BIN)/golangci-lint
# Минимальная версия Go
MIN_GO_VERSION = 1.21.5
# Версия языка, на которую рассчитан репозиторий (директива go в go.mod)
GO_MOD_VERSION := $(shell awk '$$1 == "go" { print $$2; exit }' go.mod 2>/dev/null)

# Добавляет флаг для тестирования на наличие гонок
ifdef GO_RACE_DETECTOR
    FLAGS += -race
endif

##################### Проверки для запуска golang-ci #####################
# GOLANGCI_TAG выше — единственный источник правды о версии линтера, а
# .golangci.yaml написан в формате конфига этого мажора: бинарь другого мажора
# либо отвергает конфиг целиком (v2: «can't load config: unsupported version of
# the configuration»), либо молча игнорирует большую часть настроек. Поэтому
# проверки ниже требуют *точного* совпадения версии, а иначе собирают пиннутую
# версию в ./bin (см. install-lint). Тот же паттерн, что в go-project-starter
# (TOOLS-29) и go-activerecord (LIB-11), — чтобы репозитории не расходились.
#
# Версия бинаря golangci-lint, пусто если определить не удалось. Поиск первого
# semver в выводе переваривает оба формата релизов:
#   v1: "golangci-lint has version 1.64.8 built from <commit> on <date>"
#   v2: "golangci-lint has version 2.5.0 built with go1.25.1 from (...) on (...)"
golangci_version = $(shell $(1) --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -n1)
# Go-тулчейн, которым собран бинарь ("<path>: go1.26.0" -> "1.26.0"), пусто если
# неизвестен. GOTOOLCHAIN=local не даёт пробе переключать тулчейн.
golangci_go_version = $(shell GOTOOLCHAIN=local go version -m $(1) 2>/dev/null | awk 'NR==1 { sub(/^go/, "", $$2); print $$2 }')
# Непусто, когда версия $(2) известна и меньше требуемой $(1); неизвестную $(2)
# оставляем в покое, чтобы сбойная проба не загоняла сборку в цикл.
version_older = $(if $(1),$(if $(2),$(filter-out $(1),$(word 1, $(sort $(1) $(2))))))
# Бинарь непригоден, если его версия не равна точно GOLANGCI_TAG либо он собран
# тулчейном старше той версии Go, на которую рассчитан репозиторий, — второе
# роняет линтер на тайпчеке ("file requires newer Go version goX.Y").
golangci_unusable = $(if $(filter $(GOLANGCI_TAG),$(call golangci_version,$(1))),,version)$(call version_older,$(GO_MOD_VERSION),$(call golangci_go_version,$(1)))

# Проверка локальной версии бинаря
ifneq ($(wildcard $(GOLANGCI_BIN)),)
ifneq ($(call golangci_unusable,$(GOLANGCI_BIN)),)
GOLANGCI_BIN:=
endif
endif

# Проверка глобальной версии бинаря
GOLANGCI_GLOBAL_BIN:=$(shell which golangci-lint)
ifneq ($(GOLANGCI_GLOBAL_BIN),)
ifeq ($(call golangci_unusable,$(GOLANGCI_GLOBAL_BIN)),)
GOLANGCI_BIN:=$(GOLANGCI_GLOBAL_BIN)
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

# Проверяет .golangci.yaml на соответствие схеме конфига пиннутой версии линтера.
# `golangci-lint run` игнорирует незнакомые ключи, поэтому конфиг, оставшийся от
# другого мажора, продолжает «работать», молча потеряв часть настроек, — заметит
# это только эта цель. Запускать при каждой смене GOLANGCI_TAG.
# В CI-гейт не входит: схема тянется с golangci-lint.run, и недоступность сайта
# не должна красить все pull request'ы.
.PHONY: lint-config-verify
lint-config-verify: install-lint
	$(GOLANGCI_BIN) config verify --config=.golangci.yaml

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
