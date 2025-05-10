

build-docker:
	docker build -f ./build/docker/Dockerfile -t chalk:latest .

build-docker-local:
	make build-bin IN_FILE=./cmd/app/main.go OUT_FILE=./build/bin/app-docker-local OS_BUILD=linux ARCH_BUILD=amd64
	docker build -f ./build/docker/Dockerfile.local -t chalk:latest .


ARCH_BUILD=amd64
OS_BUILD=linux
IN_FILE=./cmd/app/main.go
OUT_FILE=./build/bin/app

build-bin:
	mkdir -p build/bin
	CGO_ENABLED=0 GOOS=${OS_BUILD} GOARCH=${ARCH_BUILD} go build -v -o ${OUT_FILE} ${IN_FILE}

run-local:
	go run ./cmd/app/main.go --config ./config/config.yaml --migrations ./migrations


reset-postgres: migrate-down migrate-up

migrate-up:
	go run ./cmd/migrator/main.go \
	--postgres_uri="postgres://postgres:password@localhost:5432/postgres" \
	--migrations=./migrations \
	--up

STEPS=1
migrate-down:
	go run ./cmd/migrator/main.go \
	--postgres_uri="postgres://postgres:password@localhost:5432/postgres" \
	--migrations=./migrations \
	--down ${STEPS}