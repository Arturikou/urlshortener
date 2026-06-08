SHORTENER_BIN=cmd/shortener/shortener
TEST_BIN=./shortenertest-darwin-arm64

TEST ?= TestIteration27
SERVER_PORT=8081
DSN="postgres://derkachartur:root@localhost:5432/derkachartur?sslmode=disable"
FILE_STORAGE_PATH=/tmp/short-url-db.json

build:
	cd cmd/shortener && go build -o shortener *.go

test:
	chmod +x $(TEST_BIN)
	$(TEST_BIN) \
		-test.v \
		-test.run=^$(TEST)$$ \
		-binary-path=$(SHORTENER_BIN) \
		-server-port=$(SERVER_PORT) \
		-source-path=. \
		-file-storage-path=$(FILE_STORAGE_PATH) \
		-database-dsn=$(DSN)

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/shortener/v1/shortener.proto
