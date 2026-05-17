#!/usr/bin/env bash
set -euo pipefail

APP_NAME="weazlfeed"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_ROOT="${WEAZLFEED_HOME:-"$HOME/.weazlfeed"}"
BIN_DIR="$INSTALL_ROOT/bin"
BIN_PATH="$BIN_DIR/$APP_NAME"
GO_CACHE="${GOCACHE:-"$REPO_ROOT/.gocache"}"
GO_MOD_CACHE="${GOMODCACHE:-"$REPO_ROOT/.gomodcache"}"

go_version_number() {
  go version | awk '{print $3}' | sed 's/^go//' | cut -d. -f1,2
}

version_at_least() {
  local current="$1" required="$2"
  local current_major="${current%%.*}" current_minor="${current#*.}"
  local required_major="${required%%.*}" required_minor="${required#*.}"
  [[ "$current_major" =~ ^[0-9]+$ && "$current_minor" =~ ^[0-9]+$ ]] || return 1
  [[ "$required_major" =~ ^[0-9]+$ && "$required_minor" =~ ^[0-9]+$ ]] || return 1
  (( current_major > required_major )) && return 0
  (( current_major == required_major && current_minor >= required_minor ))
}

check_go_version() {
  if ! command -v go >/dev/null 2>&1; then
    echo "Go is required to build $APP_NAME, but it was not found on PATH." >&2
    exit 1
  fi
  local required current
  required="$(awk '/^go / {print $2; exit}' "$REPO_ROOT/go.mod" | cut -d. -f1,2)"
  current="$(go_version_number)"
  if ! version_at_least "$current" "$required"; then
    echo "Go $required or newer is required to build $APP_NAME. Found Go $current." >&2
    exit 1
  fi
}

choose_profile() {
  case "$(basename "${SHELL:-}")" in
    zsh) echo "$HOME/.zshrc" ;;
    bash)
      if [[ -f "$HOME/.bashrc" ]]; then echo "$HOME/.bashrc"; else echo "$HOME/.profile"; fi
      ;;
    fish) echo "" ;;
    *) echo "$HOME/.profile" ;;
  esac
}

check_go_version
mkdir -p "$BIN_DIR" "$GO_CACHE" "$GO_MOD_CACHE"

echo "Building $APP_NAME..."
(
  cd "$REPO_ROOT"
  GOCACHE="$GO_CACHE" GOMODCACHE="$GO_MOD_CACHE" go build -buildvcs=false -o "$BIN_PATH" ./cmd/weazlfeed
)
chmod 0755 "$BIN_PATH"

path_line='export PATH="$HOME/.weazlfeed/bin:$PATH"'
marker_begin="# >>> weazlfeed path >>>"
marker_end="# <<< weazlfeed path <<<"

if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
  profile="$(choose_profile)"
  if [[ -n "$profile" ]]; then
    touch "$profile"
    if ! grep -Fq "$marker_begin" "$profile"; then
      {
        echo ""
        echo "$marker_begin"
        echo "$path_line"
        echo "$marker_end"
      } >> "$profile"
      echo "Added $BIN_DIR to PATH in $profile"
    fi
  else
    echo "Fish shell detected. Add this to your fish config:"
    echo "set -gx PATH $BIN_DIR \$PATH"
  fi
fi

echo "Installed $APP_NAME to $BIN_PATH"
echo ""
echo "Configuring local model provider..."
(
  cd "$REPO_ROOT"
  GOCACHE="$GO_CACHE" GOMODCACHE="$GO_MOD_CACHE" go run -buildvcs=false ./cmd/weazlfeed-setup
)

if [[ "${WEAZLFEED_SKIP_LAUNCH:-}" == "1" ]]; then
  echo "Skipping first launch because WEAZLFEED_SKIP_LAUNCH=1"
else
  echo "Launching $APP_NAME..."
  exec "$BIN_PATH"
fi
