BINARY_NAME ?= clyst
OUTPUT_DIR ?= dist
PLATFORMS ?= darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64
GO ?= go
GOFLAGS ?=

.PHONY: build build-all clean tidy

build:
	@mkdir -p $(OUTPUT_DIR)
	$(GO) build $(GOFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME) .

build-all:
	@mkdir -p $(OUTPUT_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform#*/}; \
		EXT=""; \
		if [ "$$OS" = "windows" ]; then EXT=".exe"; fi; \
		OUT="$(OUTPUT_DIR)/$(BINARY_NAME)-$$OS-$$ARCH$$EXT"; \
		echo ">> building $$OUT"; \
		GOOS=$$OS GOARCH=$$ARCH $(GO) build $(GOFLAGS) -o "$$OUT" . || exit $$?; \
	done

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(OUTPUT_DIR)
