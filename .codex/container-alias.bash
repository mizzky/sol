# リポジトリのルートで source .codex/container-alias.bash を実行する。
_sol_coffeesys_codex_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd -P)"

_sol_coffeesys_codex() {
  case "$PWD/" in
    "$_sol_coffeesys_codex_root/"*)
      HERDR_AGENT=codex docker exec -it \
        -u vscode \
        -w /workspaces/sol_coffeesys \
        sol_coffeesys_devcontainer-app-1 codex "$@"
      ;;
    *)
      command codex "$@"
      ;;
  esac
}

alias codex='_sol_coffeesys_codex'
