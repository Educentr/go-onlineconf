.PHONY: test
test:
	@go test ./... -v -count=1 -coverprofile .cover ./...

.PHONY: race
race:
	@go test ./... -v -race -parallel=10

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: install-lint
install-lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1

.PHONY: coverage
coverage:
	@cat .cover | grep -v "_mock.go" | grep -v "_gen.go" | grep -v "module.go" | grep -v "test" > .cover_shrink
	@rm .cover
	@go tool cover -html=.cover_shrink -o coverage.html
	@go tool cover -func .cover_shrink | grep "total:"
	@rm .cover_shrink


.PHONY: txt_coverage
txt_coverage:
	go tool covdata textfmt -i=coverage -o profile.txt && go tool cover -func=profile.txt


bin/: ; mkdir -p $@
bin/mockgen: | bin/
	GOBIN="$(realpath $(dir $@))" go install github.com/golang/mock/mockgen@v1.6.0
