package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var safeIDRegex = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func sanitizeVaultID(id string) string {
	s := safeIDRegex.ReplaceAllString(strings.TrimSpace(id), "_")
	if s == "" {
		return "default"
	}
	return s
}

// GetProjectVaultDir 获取指定项目的服务端安全隔离凭据目录 (~/.tim-agent/vault/<project_id>)
// 该目录严格位于用户主目录下，绝对不在项目 WorkDir 内，彻底杜绝误提交到 Git 的风险。
func GetProjectVaultDir(projectID string) string {
	safeProjID := sanitizeVaultID(projectID)
	if customBase := strings.TrimSpace(os.Getenv("TIM_AGENT_VAULT_DIR")); customBase != "" {
		return filepath.Join(customBase, safeProjID)
	}
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		dir := filepath.Join(home, ".tim-agent", "vault", safeProjID)
		if err := os.MkdirAll(dir, 0700); err == nil {
			return dir
		}
	}
	return filepath.Join(os.TempDir(), ".tim-agent-vault", safeProjID)
}

func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// SyncProjectVault 将项目的多应用、多环境部署凭据写入服务端隔离的 ~/.tim-agent/vault/<project_id>/ 目录 (0700/0600 权限)
func SyncProjectVault(proj ProjectConfig) error {
	if proj.ID == "" {
		return nil
	}
	vaultDir := GetProjectVaultDir(proj.ID)
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		return err
	}
	_ = os.Chmod(vaultDir, 0700)

	// 生成各部署应用与各环境的隔离凭据文件 (.env, 0600) 与免明文执行器脚本 (.sh, 0700)
	for _, app := range proj.DeployApps {
		for _, env := range app.Envs {
			if err := writeDeployVaultFiles(vaultDir, proj, app, env); err != nil {
				return err
			}
		}
	}

	return nil
}

func GetDeployVaultFilePaths(projectID, appID, envID string) (envPath string, scriptPath string) {
	vaultDir := GetProjectVaultDir(projectID)
	prefix := fmt.Sprintf("deploy-%s-%s", sanitizeVaultID(appID), sanitizeVaultID(envID))
	return filepath.Join(vaultDir, prefix+".env"), filepath.Join(vaultDir, prefix+".sh")
}

func writeDeployVaultFiles(vaultDir string, proj ProjectConfig, app DeployAppConfig, env DeployEnvConfig) error {
	envPath, scriptPath := GetDeployVaultFilePaths(proj.ID, app.ID, env.ID)
	askpassPath := filepath.Join(vaultDir, fmt.Sprintf(".ssh-askpass-%s-%s.sh", sanitizeVaultID(app.ID), sanitizeVaultID(env.ID)))
	ignoreFilePath := filepath.Join(vaultDir, fmt.Sprintf("deploy-%s-%s.ignore", sanitizeVaultID(app.ID), sanitizeVaultID(env.ID)))
	keyFilePath := filepath.Join(vaultDir, fmt.Sprintf("deploy-%s-%s.pem", sanitizeVaultID(app.ID), sanitizeVaultID(env.ID)))

	port := env.Port
	if port <= 0 {
		if env.Method == "sftp_ftp" {
			port = 21
		} else {
			port = 22
		}
	}

	// 如果用户直接粘贴了 SSH 证书文本内容（或把证书文本贴在了 SSHKeyPath 中），自动写入 0600 权限的 .pem 密钥文件
	effectiveSSHKeyPath := strings.TrimSpace(env.SSHKeyPath)
	rawKeyText := strings.TrimSpace(env.SSHKeyContent)
	if rawKeyText == "" && strings.Contains(effectiveSSHKeyPath, "PRIVATE KEY") {
		rawKeyText = effectiveSSHKeyPath
		effectiveSSHKeyPath = ""
	}
	if rawKeyText != "" {
		if err := os.WriteFile(keyFilePath, []byte(rawKeyText+"\n"), 0600); err != nil {
			return err
		}
		_ = os.Chmod(keyFilePath, 0600)
		effectiveSSHKeyPath = keyFilePath
	}

	// 写入忽略文件列表 (.ignore)
	var cleanIgnoreLines []string
	for _, line := range strings.Split(env.IgnoreFiles, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanIgnoreLines = append(cleanIgnoreLines, trimmed)
		}
	}
	_ = os.WriteFile(ignoreFilePath, []byte(strings.Join(cleanIgnoreLines, "\n")+"\n"), 0600)

	// 1. 写入 0600 部署环境凭据与参数文件 (.env)
	var lines []string
	lines = append(lines, "# Tim-Agent 服务端隔离部署配置文件 (0600 权限，严禁明文打印或拷贝到项目代码仓库)")
	lines = append(lines, "set +x")
	lines = append(lines, fmt.Sprintf("export DEPLOY_PROJECT_NAME=%s", shellSingleQuote(proj.Name)))
	lines = append(lines, fmt.Sprintf("export DEPLOY_PROJECT_WORKDIR=%s", shellSingleQuote(proj.WorkDir)))
	lines = append(lines, fmt.Sprintf("export DEPLOY_APP_ID=%s", shellSingleQuote(app.ID)))
	lines = append(lines, fmt.Sprintf("export DEPLOY_APP_NAME=%s", shellSingleQuote(app.Name)))
	lines = append(lines, fmt.Sprintf("export DEPLOY_APP_TYPE=%s", shellSingleQuote(app.Type)))
	lines = append(lines, fmt.Sprintf("export DEPLOY_ENV_ID=%s", shellSingleQuote(env.ID)))
	lines = append(lines, fmt.Sprintf("export DEPLOY_ENV_NAME=%s", shellSingleQuote(env.EnvName)))
	lines = append(lines, fmt.Sprintf("export DEPLOY_METHOD=%s", shellSingleQuote(env.Method)))
	lines = append(lines, fmt.Sprintf("export DEPLOY_HOST=%s", shellSingleQuote(env.Host)))
	lines = append(lines, fmt.Sprintf("export DEPLOY_PORT=%d", port))
	lines = append(lines, fmt.Sprintf("export DEPLOY_USER=%s", shellSingleQuote(env.Username)))
	lines = append(lines, fmt.Sprintf("export DEPLOY_AUTH_TYPE=%s", shellSingleQuote(env.AuthType)))
	lines = append(lines, fmt.Sprintf("export DEPLOY_PASSWORD=%s", shellSingleQuote(env.Password)))
	lines = append(lines, fmt.Sprintf("export DEPLOY_SSH_KEY=%s", shellSingleQuote(effectiveSSHKeyPath)))
	lines = append(lines, fmt.Sprintf("export DEPLOY_REMOTE_DIR=%s", shellSingleQuote(env.RemoteDir)))
	lines = append(lines, fmt.Sprintf("export DEPLOY_IGNORE_FILE=%s", shellSingleQuote(ignoreFilePath)))

	if err := os.WriteFile(envPath, []byte(strings.Join(lines, "\n")+"\n"), 0600); err != nil {
		return err
	}
	_ = os.Chmod(envPath, 0600)

	// 2. 写入 SSH_ASKPASS 辅助脚本 (支持无 sshpass 时通过 SSH_ASKPASS 免交互输入密码)
	askpassScript := fmt.Sprintf(`#!/usr/bin/env bash
set +x
. %s
printf '%%s\n' "$DEPLOY_PASSWORD"
`, shellSingleQuote(envPath))
	if err := os.WriteFile(askpassPath, []byte(askpassScript), 0700); err != nil {
		return err
	}
	_ = os.Chmod(askpassPath, 0700)

	// 3. 写入免明文安全部署执行脚本 (.sh, 0700)
	runnerScript := fmt.Sprintf(`#!/usr/bin/env bash
set -e
set +x

# 加载服务端隔离凭据配置 (严禁开启 set -x 防止回显密码)
. %s
export SSH_ASKPASS=%s
export SSH_ASKPASS_REQUIRE=force
export DISPLAY="${DISPLAY:-:0}"

run_ssh() {
  set +x
  local cmd="$1"
  local ssh_opts="-p $DEPLOY_PORT -o StrictHostKeyChecking=accept-new"
  if [ -n "$DEPLOY_SSH_KEY" ]; then
    ssh $ssh_opts -i "$DEPLOY_SSH_KEY" "$DEPLOY_USER@$DEPLOY_HOST" "$cmd"
  elif command -v sshpass >/dev/null 2>&1 && [ -n "$DEPLOY_PASSWORD" ]; then
    SSHPASS="$DEPLOY_PASSWORD" sshpass -e ssh $ssh_opts "$DEPLOY_USER@$DEPLOY_HOST" "$cmd"
  else
    ssh $ssh_opts "$DEPLOY_USER@$DEPLOY_HOST" "$cmd"
  fi
}

run_upload() {
  set +x
  local src="${1:-.}"
  local dst="${2:-$DEPLOY_REMOTE_DIR}"
  echo "[TimAgent Deploy] 正在上传 [$src] -> 目标目录 [$dst] ..."
  if [ "$DEPLOY_METHOD" = "sftp_ftp" ]; then
    curl -sS --ftp-create-dirs -u "$DEPLOY_USER:$DEPLOY_PASSWORD" -T "$src" "ftp://$DEPLOY_HOST:$DEPLOY_PORT/$dst/"
  elif command -v rsync >/dev/null 2>&1 && [ -f "$DEPLOY_IGNORE_FILE" ]; then
    local ssh_cmd="ssh -p $DEPLOY_PORT -o StrictHostKeyChecking=accept-new"
    if [ -n "$DEPLOY_SSH_KEY" ]; then
      ssh_cmd="$ssh_cmd -i $DEPLOY_SSH_KEY"
    fi
    if command -v sshpass >/dev/null 2>&1 && [ -n "$DEPLOY_PASSWORD" ] && [ -z "$DEPLOY_SSH_KEY" ]; then
      SSHPASS="$DEPLOY_PASSWORD" sshpass -e rsync -az --exclude-from="$DEPLOY_IGNORE_FILE" -e "$ssh_cmd" "$src" "$DEPLOY_USER@$DEPLOY_HOST:$dst"
    else
      rsync -az --exclude-from="$DEPLOY_IGNORE_FILE" -e "$ssh_cmd" "$src" "$DEPLOY_USER@$DEPLOY_HOST:$dst"
    fi
  else
    local scp_opts="-P $DEPLOY_PORT -r -o StrictHostKeyChecking=accept-new"
    if [ -n "$DEPLOY_SSH_KEY" ]; then
      scp $scp_opts -i "$DEPLOY_SSH_KEY" "$src" "$DEPLOY_USER@$DEPLOY_HOST:$dst"
    elif command -v sshpass >/dev/null 2>&1 && [ -n "$DEPLOY_PASSWORD" ]; then
      SSHPASS="$DEPLOY_PASSWORD" sshpass -e scp $scp_opts "$src" "$DEPLOY_USER@$DEPLOY_HOST:$dst"
    else
      scp $scp_opts "$src" "$DEPLOY_USER@$DEPLOY_HOST:$dst"
    fi
  fi
}

ACTION="${1:-upload}"
case "$ACTION" in
  upload)
    cd "$DEPLOY_PROJECT_WORKDIR"
    run_upload "$2" "$3"
    ;;
  remote|ssh)
    shift || true
    CMD_TO_RUN="$*"
    if [ -n "$CMD_TO_RUN" ]; then
      run_ssh "$CMD_TO_RUN"
    fi
    ;;
  *)
    cd "$DEPLOY_PROJECT_WORKDIR"
    run_upload "$1" "$2"
    ;;
esac
`, shellSingleQuote(envPath), shellSingleQuote(askpassPath))

	if err := os.WriteFile(scriptPath, []byte(runnerScript), 0700); err != nil {
		return err
	}
	_ = os.Chmod(scriptPath, 0700)
	return nil
}

// MaskHost 对服务器地址进行轻度脱敏展示（仅用于前端 UI 摘要提示）
func MaskHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return "本地环境 (Local)"
	}
	parts := strings.Split(host, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s.%s.*.%s", parts[0], parts[1], parts[3])
	}
	return host
}

// BuildDeployPrompt 根据项目选定的部署应用与目标环境，生成零明文泄露风险的 AI 部署提示词
func BuildDeployPrompt(proj ProjectConfig, app DeployAppConfig, env DeployEnvConfig) string {
	envPath, scriptPath := GetDeployVaultFilePaths(proj.ID, app.ID, env.ID)

	var details []string
	if app.Type != "" {
		details = append(details, fmt.Sprintf("应用类型: `%s`", app.Type))
	}
	if env.Method != "" {
		details = append(details, fmt.Sprintf("部署方式: `%s`", env.Method))
	}
	if env.RemoteDir != "" {
		details = append(details, fmt.Sprintf("远程目标目录: `%s`", env.RemoteDir))
	}

	var ignoreList []string
	for _, line := range strings.Split(env.IgnoreFiles, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			ignoreList = append(ignoreList, s)
		}
	}
	if len(ignoreList) > 0 {
		details = append(details, fmt.Sprintf("忽略文件: `%s`", strings.Join(ignoreList, ", ")))
	}

	summaryStr := ""
	if len(details) > 0 {
		summaryStr = "（" + strings.Join(details, "，") + "）"
	}

	customPromptStr := ""
	if cp := strings.TrimSpace(env.CustomPrompt); cp != "" {
		// 将多行换行整理为空格或分号，避免发给终端时被截断为多次回车
		compactCP := strings.Join(strings.Fields(strings.ReplaceAll(cp, "\n", "；")), " ")
		customPromptStr = fmt.Sprintf(" 【自定义部署要求】：%s。", compactCP)
	}

	return fmt.Sprintf(
		"请帮我将当前项目「%s」下的应用【%s】部署到【%s】环境%s。%s "+
			"【安全部署工具】：服务器账号密码与连接配置已隔离保存在 `%s` 与免密执行脚本 `%s` 中（上传用 `%s upload <本地路径> [远程路径]`，远程执行命令用 `%s ssh \"<命令>\"`）。"+
			"⚠️ 严禁使用 cat/echo 打印 `%s` 中的明文密码，严禁将凭据复制到项目代码目录中！",
		proj.Name, app.Name, env.EnvName, summaryStr, customPromptStr,
		envPath, scriptPath, scriptPath, scriptPath, envPath,
	)
}
