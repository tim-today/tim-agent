.PHONY: all build app build-mac-arm64 build-mac-amd64 build-win build-linux dist clean

BINARY_NAME=tim-agent
BIN_DIR=bin

all: build app

# 编译当前平台命令行程序
build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) .

# 生成本机 macOS GUI 应用程序包 (bin/TimAgent.app，鼠标双击直接打开，无终端黑框)
app: build
	./$(BIN_DIR)/$(BINARY_NAME) -pack-mac-app $(BIN_DIR)/TimAgent.app -pack-bin $(BIN_DIR)/$(BINARY_NAME)

# 一键安装到本机系统桌面与应用程序目录 (/Applications/TimAgent.app 或桌面图标)
install-app: build
	./$(BIN_DIR)/$(BINARY_NAME) -install-app

# macOS Apple Silicon (M1/M2/M3/M4) CLI + GUI .app
build-mac-arm64: build
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 .
	./$(BIN_DIR)/$(BINARY_NAME) -pack-mac-app $(BIN_DIR)/TimAgent-arm64.app -pack-bin $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64
	cd $(BIN_DIR) && rm -f TimAgent-macos-arm64.app.zip && zip -r -q TimAgent-macos-arm64.app.zip TimAgent-arm64.app && rm -rf TimAgent-arm64.app

# macOS Intel CLI + GUI .app
build-mac-amd64: build
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 .
	./$(BIN_DIR)/$(BINARY_NAME) -pack-mac-app $(BIN_DIR)/TimAgent-amd64.app -pack-bin $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64
	cd $(BIN_DIR) && rm -f TimAgent-macos-amd64.app.zip && zip -r -q TimAgent-macos-amd64.app.zip TimAgent-amd64.app && rm -rf TimAgent-amd64.app

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

# 编译所有支持的平台与CPU架构 (CLI + GUI Application)
dist: build-mac-arm64 build-mac-amd64 build-win build-linux app
	@echo "所有目标平台 CLI 与 GUI 应用程序构建完成，存放于 $(BIN_DIR)/ 目录"

clean:
	rm -rf $(BIN_DIR)
