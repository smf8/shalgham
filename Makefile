#@IgnoreInspection BashAddShebang

export APP=shalgham

export ROOT=$(realpath $(dir $(lastword $(MAKEFILE_LIST))))

export LDFLAGS="-w -s"

export POSTGRES_ADDRESS=127.0.0.1:54320
export POSTGRES_DATABASE=shalgham
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=pass
export POSTGRES_DSN="postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_ADDRESS)/$(POSTGRES_DATABASE)?sslmode=disable"
all: format lint build

############################################################
# Migrations
############################################################

migrate-create:
	migrate create -ext sql -dir ./migrations $(NAME)

migrate-up:
	migrate -verbose  -path ./migrations -database $(POSTGRES_DSN) up

migrate-down:
	 migrate -path ./migrations -database $(POSTGRES_DSN) down

migrate-reset:
	 migrate -path ./migrations -database $(POSTGRES_DSN) drop

migrate-install:
	which migrate || GO111MODULE=off go get -tags 'postgres' -v -u github.com/golang-migrate/migrate/cmd/migrate

############################################################
# Build and Run
############################################################

build:
	CGO_ENABLED=1 go build -ldflags $(LDFLAGS)  .

build-static:
	CGO_ENABLED=1 go build -v -o $(APP) -a -installsuffix cgo -ldflags $(LDFLAGS) .

install:
	CGO_ENABLED=1 go install -ldflags $(LDFLAGS)

# Please do not use `make run` on production. There is a performance hit due to existence of -race flag.
run-server:
	go run -race -ldflags $(LDFLAGS) . server
run-client:
	go run -race -ldflags $(LDFLAGS) . client localhost:1234

############################################################
# Format and Lint
############################################################

check-formatter:
	which goimports || GO111MODULE=off go get -u golang.org/x/tools/cmd/goimports

format: check-formatter
	find $(ROOT) -type f -name "*.go" -not -path "$(ROOT)/vendor/*" | xargs -n 1 -I R goimports -w R
	find $(ROOT) -type f -name "*.go" -not -path "$(ROOT)/vendor/*" | xargs -n 1 -I R gofmt -s -w R

check-linter:
	which golangci-lint || GO111MODULE=off curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.23.1

lint: check-linter
	golangci-lint run --deadline 10m $(ROOT)/...

############################################################
# Test
############################################################

test:
	go test -v -race -p 1 ./...

ci-test:
	go test -v -race -p 1 -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -func coverage.txt

############################################################
# Development Environment
############################################################

up:
	docker-compose up -d

down:
	docker-compose down
