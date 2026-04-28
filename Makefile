SHORTENER_BIN=cmd/shortener/shortener
TEST_BIN=./shortenertest-darwin-arm64

# Значение по умолчанию для одиночного запуска
TEST ?= TestIteration16
SERVER_PORT=8081
DSN="postgres://derkachartur:root@localhost:5432/derkachartur?sslmode=disable"
FILE_STORAGE_PATH=/tmp/short-url-db.json

build:
	cd cmd/shortener && go build -o shortener *.go

# Одиночный запуск (как и было)
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

# Запуск всех тестов от 1 до 17
test-all: build
	chmod +x $(TEST_BIN)
	@for i in {1..17}; do \
		echo "--- Running TestIteration$$i ---"; \
		$(TEST_BIN) \
			-test.v \
			-test.run=^TestIteration$$i$$ \
			-binary-path=$(SHORTENER_BIN) \
			-server-port=$(SERVER_PORT) \
			-source-path=. \
			-file-storage-path=$(FILE_STORAGE_PATH) \
			-database-dsn=$(DSN) || exit 1; \
	done