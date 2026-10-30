.PHONY: all build app install-app build-mac-arm64 build-mac-amd64 build-mac-universal build-win build-linux dist clean

BINARY_NAME=tim-agent
BIN_DIR=bin
MACOS_ENV=MACOSX_DEPLOYMENT_TARGET=11.0 CGO_CFLAGS="-mmacosx-version-min=11.0" CGO_LDFLAGS="-mmacosx-version-min=11.0"

all: build app

# 编译当前平台命令行程序 (macOS 默认开启 11.0+ 向下兼容)
build:
	mkdir -p $(BIN_DIR)
	$(MACOS_ENV) go build -o $(BIN_DIR)/$(BINARY_NAME) .

# 生成本机 macOS Universal 通用架构 GUI 应用程序包 (bin/TimAgent.app，同时兼容 Intel x86_64 与 Apple Silicon arm64 及 macOS 11.0+)
app: build-mac-universal
	./$(BIN_DIR)/$(BINARY_NAME) -pack-mac-app $(BIN_DIR)/TimAgent.app -pack-bin $(BIN_DIR)/$(BINARY_NAME)-darwin-universal
	cd $(BIN_DIR) && rm -f TimAgent-macos-universal.app.zip && zip -r -q TimAgent-macos-universal.app.zip TimAgent.app

# 一键安装 Universal 通用架构 TimAgent.app 到本机系统桌面与 /Applications
install-app: app
	./$(BIN_DIR)/TimAgent.app/Contents/MacOS/tim-agent -install-app

# macOS Apple Silicon (M1/M2/M3/M4)
build-mac-arm64:
	mkdir -p $(BIN_DIR)
	$(MACOS_ENV) CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 .

# macOS Intel (x86_64)
build-mac-amd64:
	mkdir -p $(BIN_DIR)
	$(MACOS_ENV) CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 .

# macOS Universal (Intel x86_64 + Apple Silicon arm64 双架构合一)
build-mac-universal: build build-mac-arm64 build-mac-amd64
	lipo -create -output $(BIN_DIR)/$(BINARY_NAME)-darwin-universal $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64

# Windows 64-bit (同时生成命令行版与无黑框鼠标双击 GUI 版)
build-win:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H=windowsgui" -o $(BIN_DIR)/TimAgent-windows-amd64-gui.exe .

# Linux (AMD64 & ARM64 命令行版 + 桌面 GUI 应用包)
build-linux: build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 .
	./$(BIN_DIR)/$(BINARY_NAME) -pack-linux-app $(BIN_DIR)/TimAgent-linux-amd64-gui -pack-bin $(BIN_DIR)/$(BINARY_NAME)-linux-amd64
	cd $(BIN_DIR) && tar -czf TimAgent-linux-amd64-gui.tar.gz TimAgent-linux-amd64-gui && rm -rf TimAgent-linux-amd64-gui
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 .
	./$(BIN_DIR)/$(BINARY_NAME) -pack-linux-app $(BIN_DIR)/TimAgent-linux-arm64-gui -pack-bin $(BIN_DIR)/$(BINARY_NAME)-linux-arm64
	cd $(BIN_DIR) && tar -czf TimAgent-linux-arm64-gui.tar.gz TimAgent-linux-arm64-gui && rm -rf TimAgent-linux-arm64-gui

# 编译所有支持的平台与CPU架构 (CLI + Universal GUI Application)
dist: app build-win build-linux
	@echo "所有目标平台 CLI 与 Universal GUI 应用程序构建完成，存放于 $(BIN_DIR)/ 目录"

clean:
	rm -rf $(BIN_DIR)
