GITHUB_SHA ?= ct-logic-api-document
SRC = `go list -f {{.Dir}} ./... | grep -v /vendor/ `

fmt:
	@echo "==> Formatting source code..."
	@echo "==> Running gofumpt..."
	@gofumpt -w $(SRC)
	@echo "==> Running goimports-reviser..."
	@go list -f {{.Dir}} ./... | grep -v /vendor/ | sed 's|$(SRC_ROOT)|.|g' | xargs -I {} goimports-reviser -rm-unused -company-prefixes $(COMPANY_REPO) {}
	@git diff --quiet

lint:
	@echo "==> Running lint check..."
	@golangci-lint --config setup/.golangci.yml run
	@go vet -tags=$(TAGS) `go list -f {{.Dir}} ./... | grep -v /vendor/`

test:
	@echo "==> Running test"
	@go clean -testcache
	go test -vet=off ./... -p 1 -race -cover -coverprofile=coverage.out

dev-up:
	@docker compose \
		-f setup/docker-compose.dev.yml \
		-p $(GITHUB_SHA) up --build -d \
		--remove-orphans

dev-down:
	@docker compose \
		-f setup/docker-compose.dev.yml \
		-p $(GITHUB_SHA) down \
		-v --rmi local

dev:
	go mod tidy
	go run main.go service

start-swagger:
	docker run -e QUERY_CONFIG_ENABLED='true' -p 80:8080 swaggerapi/swagger-ui

stop-swagger:
	docker stop $(docker ps -a -q --filter ancestor=swaggerapi/swagger-ui)


fetch-data-local:
	echo "==> Fetching data from local"
	go run main.go cronjob fetch_data_from_local 

build-structure:
	echo "==> Building structure"
	go run main.go cronjob build_structure 


.PHONY: test dev-up dev-down fmt lint dev start-swagger fetch-data-local build-structure