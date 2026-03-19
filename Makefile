# Include variables from the .envrc file
include .envrc

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]



# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api

## run/frontend: run the frontend application
.PHONY: run/frontend
run/frontend:
	cd frontend && npm run dev

## install/deps: install the deps
.PHONY: install/deps
install/deps:
	cd frontend && npm install
	cd ..

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${KICK_ALERT_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Crating a new migration files for ${name}'
	cd ./sql/schema/ && goose -s create ${name} sql
	cd ../..

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Runing up migration'
	cd ./sql/schema && goose postgres ${KICK_ALERT_DB_DSN} up
	cd ../..


## db/migrations/down: apply all donw database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo 'Runing down migration'
	cd ./sql/schema && goose postgres ${KICK_ALERT_DB_DSN} down
	cd ../..


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: tidy module dependecies and format all .go files
.PHONY: tidy
tidy:
	@echo 'Tidying module dependecies...'
	go mod tidy 
	@echo 'verifying and vendoring module dependecies...'
	go mod verify
	go mod vendor
	@echo 'Formating .go files'
	go fmt ./...

## audit: run quality control checks
.PHONY: audit
audit:
	@echo 'Checking module dependencies...'
	go mod tidy -diff
	go mod verify
	@echo 'Vetting code...'
	go vet ./...
	go tool staticcheck ./...
	@echo 'Runing Tests...'
	go test -race -vet=off ./...

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags='-s'  -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api