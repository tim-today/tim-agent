.PHONY: all build build-mac build-mac-arm64 build-mac-amd64 build-win clean

BINARY_NAME=tim-agent
BIN_DIR=bin

all: build

build:
	go build -o $(BIN_DIR)/$(BINARY_NAME) .

# macOS Apple Silicon (M1/M2/M3/M4)
build-mac-arm64:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 .

# macOS Intel
build-mac-amd64:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 .

# Windows 64-bit
build-win:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -tags nocgo -o $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# 编译所有支持的平台与CPU架构
dist: build-mac-arm64 build-mac-amd64 build-win
	@echo "所有目标平台构建完成，存放于 $(BIN_DIR)/ 目录"

clean:
	rm -rf $(BIN_DIR)
