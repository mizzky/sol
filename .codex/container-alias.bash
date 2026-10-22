# リポジトリのルートで source .codex/container-alias.bash を実行する。
_sol_coffeesys_codex_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd -P)"

_sol_coffeesys_codex() {
  case "$PWD/" in
    "$_sol_coffeesys_codex_root/"*)
      devcontainer up \
        --workspace-folder "$_sol_coffeesys_codex_root" || return

      HERDR_AGENT=codex devcontainer exec \
        --workspace-folder "$_sol_coffeesys_codex_root" \
        --remote-env HERDR_AGENT=codex \
        codex "$@"
      ;;
    *)
      command codex "$@"
      ;;
  esac
}

alias codex='_sol_coffeesys_codex'
