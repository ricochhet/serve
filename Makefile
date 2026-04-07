BUILD_OUTPUT=build
ASSET_PATH=assets

CUSTOM=-X 'main.buildDate=$(shell date)' -X 'main.gitHash=$(shell git rev-parse --short HEAD)' -X 'main.buildOn=$(shell go version)'
LDFLAGS=$(CUSTOM) -w -s -extldflags=-static
GO_BUILD=go build -trimpath -ldflags "$(LDFLAGS)"

APP_NAMES=serve

SERVE_PATH=./cmd/serve
SERVE_BIN_NAME=serve

PM_PATH=./cmd/pm
PM_BIN_NAME=pm

define GO_BUILD_APP
	CGO_ENABLED=1 GOOS=$(1) GOARCH=$(2) $(GO_BUILD) -o $(BUILD_OUTPUT)/$(3) $(4)
endef

.PHONY: all
all: serve

.PHONY: fmt
fmt:
	gofumpt -l -w -extra .

.PHONY: tidy
tidy:
	@echo "[main] tidy"
	go mod tidy

.PHONY: update
update:
	@echo "[main] update dependencies"
	go get -u ./...

.PHONY: lint
lint: fmt
	@echo "[main] golangci-lint"
	golangci-lint run ./... --fix

.PHONY: test
test:
	go test ./...

.PHONY: deadcode
deadcode:
	deadcode ./...

.PHONY: syso
syso:
	windres $(SERVE_PATH)/app.rc -O coff -o $(SERVE_PATH)/app.syso

.PHONY: png-to-icos
png-to-icos:
	magick $(ASSET_PATH)/win-icon.png -background none -define icon:auto-resize=256,128,64,48,32,16 $(ASSET_PATH)/win-icon.ico

.PHONY: copy-assets
copy-assets:
	cp -r $(ASSET_PATH)/* $(BUILD_OUTPUT)

.PHONY: gen-certs
gen-certs:
	mkcert localhost 127.0.0.1 ::1

# ----- serve -----
.PHONY: serve
serve: serve-linux serve-linux-arm64 serve-darwin serve-darwin-arm64 serve-windows

.PHONY: serve-linux
serve-linux: fmt
	$(call GO_BUILD_APP,linux,amd64,$(SERVE_BIN_NAME)-linux,$(SERVE_PATH))

.PHONY: serve-linux-arm64
serve-linux-arm64: fmt
	$(call GO_BUILD_APP,linux,arm64,$(SERVE_BIN_NAME)-linux-arm64,$(SERVE_PATH))

.PHONY: serve-darwin
serve-darwin: fmt
	$(call GO_BUILD_APP,darwin,amd64,$(SERVE_BIN_NAME)-darwin,$(SERVE_PATH))

.PHONY: serve-darwin-arm64
serve-darwin-arm64: fmt
	$(call GO_BUILD_APP,darwin,arm64,$(SERVE_BIN_NAME)-darwin-arm64,$(SERVE_PATH))

.PHONY: serve-windows
serve-windows: fmt copy-assets
	$(call GO_BUILD_APP,windows,amd64,$(SERVE_BIN_NAME).exe,$(SERVE_PATH))

# ----- PM -----
.PHONY: pm
pm: pm-linux pm-linux-arm64 pm-darwin pm-darwin-arm64 pm-windows

.PHONY: pm-linux
pm-linux: fmt
	$(call GO_BUILD_APP,linux,amd64,$(PM_BIN_NAME)-linux,$(PM_PATH))

.PHONY: pm-linux-arm64
pm-linux-arm64: fmt
	$(call GO_BUILD_APP,linux,arm64,$(PM_BIN_NAME)-linux-arm64,$(PM_PATH))

.PHONY: pm-darwin
pm-darwin: fmt
	$(call GO_BUILD_APP,darwin,amd64,$(PM_BIN_NAME)-darwin,$(PM_PATH))

.PHONY: pm-darwin-arm64
pm-darwin-arm64: fmt
	$(call GO_BUILD_APP,darwin,arm64,$(PM_BIN_NAME)-darwin-arm64,$(PM_PATH))

.PHONY: pm-windows
pm-windows: fmt copy-assets
	$(call GO_BUILD_APP,windows,amd64,$(PM_BIN_NAME).exe,$(PM_PATH))
