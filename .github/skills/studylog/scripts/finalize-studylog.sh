#!/usr/bin/env sh
set -eu

# Purpose:
# - ログ記録指示時に、FinalizeLog チェックポイントをジャーナルへ追記する
# - ジャーナルファイルのパスを stdout に出力する（LLM がこれを読んで学習ログを生成する）

WORKSPACE_ROOT="${PWD}"

YYYY="$(date +%Y)"
MM="$(date +%m)"
DD="$(date +%d)"

LOG_DIR="${WORKSPACE_ROOT}/doc/memo/${YYYY}/${MM}"
JOURNAL_FILE="${LOG_DIR}/${DD}_journal.tmp.md"
AUTO_CHECKPOINT="${WORKSPACE_ROOT}/.github/skills/studylog/scripts/auto-checkpoint.sh"

mkdir -p "${LOG_DIR}"

FINALIZE_NOTE="${1:-ログ記録 指示による最終化}"

json_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

# FinalizeLog チェックポイントをジャーナルへ追記
if [ -x "${AUTO_CHECKPOINT}" ]; then
  ESCAPED_NOTE="$(json_escape "${FINALIZE_NOTE}")"
  printf '%s' "{\"hookEventName\":\"FinalizeLog\",\"topic\":\"ログ記録最終化\",\"userPrompt\":\"${ESCAPED_NOTE}\",\"blocker\":\"なし\",\"resolution\":\"LLMによる学習ログ生成\",\"next\":\"次回セッションへ\"}" | "${AUTO_CHECKPOINT}" >/dev/null
fi

# ジャーナルパスを出力（LLM がこのパスを読んで learning.md を生成する）
echo "${JOURNAL_FILE}"
