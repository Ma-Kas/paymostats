APP=paymostats

run:
	go run ./cmd/$(APP)

build:
	go build -o bin/$(APP) ./cmd/$(APP)

lint:
	golangci-lint run

fmt:
	gofumpt -l -w .

.PHONY: run build lint fmt
