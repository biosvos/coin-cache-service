.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -o . ./cmd/...

.PHONY: install
install:
	CGO_ENABLED=0 GOOS=linux go install ./cmd/...

.PHONY: update
update:
	go get -u -t ./...
	go mod tidy
	go mod vendor

.PHONY: validate
validate: build
	# go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	# mv $(which golangci-lint) ~/.local/bin
	golangci-lint run
	go test ./...

.PHONY: cover
cover:
	go test -shuffle on ./... --coverpkg ./... -coverprofile=c.out
	go tool cover -html="c.out"
	rm c.out
