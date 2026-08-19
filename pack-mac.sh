#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

usage() {
  cat <<'EOF'
Usage:
  ./pack-mac.sh [options]

在当前 macOS 主机上打包生成 Ant Browser.app 和 zip。

Options:
  --arch <arm64|amd64>   目标架构，默认跟随本机
  --version <ver>        版本号，默认读取 wails.json
  --open                 打包完成后打开生成的 .app
  --skip-build           跳过前端和 Wails 编译，只组装已有产物
  --skip-runtime-verify  跳过代理运行时校验
  --keep-staging         保留 publish/staging/mac 中的组装目录
  -h, --help             显示帮助

产物目录:
  publish/output/AntBrowser-<version>-macos-<arch>.app
  publish/output/AntBrowser-<version>-macos-<arch>.zip
EOF
}

if [[ "${1:-}" == "help" || "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "[ERROR] ./pack-mac.sh 只能在 macOS 上运行" >&2
  exit 1
fi

if [[ -z "${GOPROXY:-}" ]]; then
  export GOPROXY="https://goproxy.cn,direct"
fi

GO_BIN="$(go env GOPATH 2>/dev/null || true)/bin"
if [[ -n "$GO_BIN" && -d "$GO_BIN" ]]; then
  case ":$PATH:" in
    *":$GO_BIN:"*) ;;
    *) export PATH="$GO_BIN:$PATH" ;;
  esac
fi

echo "========================================"
echo "  Ant Browser macOS Pack"
echo "========================================"
echo "Root: $ROOT_DIR"
echo

exec bash "$ROOT_DIR/publish/mac/publish-mac.sh" "$@"
