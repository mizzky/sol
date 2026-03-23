#!/usr/bin/env sh
set -eu

# Purpose:
# - PreCompact hook input (JSON on stdin) を受け取る
# - 日付別の中間ジャーナルにチェックポイントを追記する
# - 解析できる項目は抽出し、解析不可でも生データ断片を残す

# Workspace root は hook の cwd を前提にする
WORKSPACE_ROOT="${PWD}"

# 入力JSONを一時ファイルに退避
TMP_INPUT="$(mktemp)"
cat > "${TMP_INPUT}"

# 日付・時刻
YYYY="$(date +%Y)"
MM="$(date +%m)"
DD="$(date +%d)"
TS="$(date +%Y-%m-%dT%H:%M:%S%z)"
CHECKPOINT_ID="cp-${TS}-$$"

# 出力先: doc/memo/YYYY/MM/DD_journal.tmp.md
LOG_DIR="${WORKSPACE_ROOT}/doc/memo/${YYYY}/${MM}"
JOURNAL_FILE="${LOG_DIR}/${DD}_journal.tmp.md"

mkdir -p "${LOG_DIR}"

# 文字列を1行に整形
oneline() {
  tr '\n' ' ' | tr '\r' ' ' | sed 's/[[:space:]][[:space:]]*/ /g' | sed 's/^ //; s/ $//'
}

# 先頭N文字に切る
clip() {
  # shellcheck disable=SC2001
  echo "$1" | sed "s/^\(.\{0,$2\}\).*/\1/"
}

# jqが使える場合はJSONから候補を抽出
EVENT_NAME="PreCompact"
TOPIC=""
QUESTION=""
BLOCKER=""
RESOLUTION=""
NEXT_STEP=""

if command -v jq >/dev/null 2>&1; then
  EVENT_NAME="$(jq -r '.hookEventName // .event // .name // "PreCompact"' "${TMP_INPUT}" 2>/dev/null || echo "PreCompact")"

  # 代表的な候補キーを幅広く拾う（存在しなければ空）
  TOPIC="$(jq -r '.topic // .task // .summary // .session.topic // ""' "${TMP_INPUT}" 2>/dev/null || true)"
  QUESTION="$(jq -r '.question // .userPrompt // .prompt // .session.question // ""' "${TMP_INPUT}" 2>/dev/null || true)"
  BLOCKER="$(jq -r '.blocker // .issue // .session.blocker // ""' "${TMP_INPUT}" 2>/dev/null || true)"
  RESOLUTION="$(jq -r '.resolution // .fix // .session.resolution // ""' "${TMP_INPUT}" 2>/dev/null || true)"
  NEXT_STEP="$(jq -r '.next // .nextStep // .session.next // ""' "${TMP_INPUT}" 2>/dev/null || true)"
fi

# 候補が空なら、生入力の断片をQuestionへ退避
RAW_SNIPPET="$(head -c 1200 "${TMP_INPUT}" | oneline)"

if [ -z "${TOPIC}" ]; then
  TOPIC="(auto) context checkpoint before compaction"
fi

if [ -z "${QUESTION}" ]; then
  QUESTION="$(clip "${RAW_SNIPPET}" 300)"
  if [ -z "${QUESTION}" ]; then
    QUESTION="(no prompt text captured)"
  fi
fi

if [ -z "${BLOCKER}" ]; then
  BLOCKER="(none captured)"
fi

if [ -z "${RESOLUTION}" ]; then
  RESOLUTION="(pending)"
fi

if [ -z "${NEXT_STEP}" ]; then
  NEXT_STEP="コンパクション後に未解決事項の確認と継続"
fi

# Markdown追記
{
  echo ""
  echo "## JournalCheckpoint ${TS}"
  echo "- CheckpointId: ${CHECKPOINT_ID}"
  echo "- Event: ${EVENT_NAME}"
  echo "- Topic: ${TOPIC}"
  echo "- Question: ${QUESTION}"
  echo "- Blocker: ${BLOCKER}"
  echo "- Resolution: ${RESOLUTION}"
  echo "- Next: ${NEXT_STEP}"
  echo ""
  echo "### RawHookInputSnippet"
  echo ""
  echo "${RAW_SNIPPET}"
  echo ""
} >> "${JOURNAL_FILE}"

# 後始末
rm -f "${TMP_INPUT}"

# Hook実行継続
# stdoutにJSONを返せる実装にしておく（なくても動く環境はある）
echo '{"continue": true}'