#!/usr/bin/env sh
set -eu

# Purpose:
# - Hook input (JSON on stdin) を受け取る
# - 日付別の中間ジャーナルにチェックポイントを追記する
# - transcript_path があれば永続化用スナップショットを保存する
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
TRANSCRIPT_DIR="${LOG_DIR}/transcripts"
INDEX_FILE="${LOG_DIR}/${DD}_transcript_index.tsv"

mkdir -p "${LOG_DIR}"
mkdir -p "${TRANSCRIPT_DIR}"

# 文字列を1行に整形
oneline() {
  tr '\n' ' ' | tr '\r' ' ' | sed 's/[[:space:]][[:space:]]*/ /g' | sed 's/^ //; s/ $//'
}

# 先頭N文字に切る
clip() {
  # shellcheck disable=SC2001
  echo "$1" | sed "s/^\(.\{0,$2\}\).*/\1/"
}

safe_token() {
  printf '%s' "$1" | tr -c 'A-Za-z0-9._-' '_' | sed 's/^$/unknown/'
}

to_workspace_relative() {
  case "$1" in
    "${WORKSPACE_ROOT}"/*) printf '%s' "${1#${WORKSPACE_ROOT}/}" ;;
    *) printf '%s' "$1" ;;
  esac
}

hash_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
    return
  fi

  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
    return
  fi

  wc -c < "$1" | awk '{print "bytes-"$1}'
}

# jqが使える場合はJSONから候補を抽出
EVENT_NAME="PreCompact"
HOOK_TIMESTAMP=""
SESSION_ID=""
TRIGGER=""
TRANSCRIPT_PATH=""
TOPIC=""
QUESTION=""
BLOCKER=""
RESOLUTION=""
NEXT_STEP=""

if command -v jq >/dev/null 2>&1; then
  EVENT_NAME="$(jq -r '.hookEventName // .event // .name // "PreCompact"' "${TMP_INPUT}" 2>/dev/null || echo "PreCompact")"
  HOOK_TIMESTAMP="$(jq -r '.timestamp // ""' "${TMP_INPUT}" 2>/dev/null || true)"
  SESSION_ID="$(jq -r '.sessionId // .session_id // ""' "${TMP_INPUT}" 2>/dev/null || true)"
  TRIGGER="$(jq -r '.trigger // ""' "${TMP_INPUT}" 2>/dev/null || true)"
  TRANSCRIPT_PATH="$(jq -r '.transcript_path // .transcriptPath // ""' "${TMP_INPUT}" 2>/dev/null || true)"

  # 代表的な候補キーを幅広く拾う（存在しなければ空）
  TOPIC="$(jq -r '.topic // .task // .summary // .session.topic // ""' "${TMP_INPUT}" 2>/dev/null || true)"
  QUESTION="$(jq -r '.question // .userPrompt // .prompt // .session.question // ""' "${TMP_INPUT}" 2>/dev/null || true)"
  BLOCKER="$(jq -r '.blocker // .issue // .session.blocker // ""' "${TMP_INPUT}" 2>/dev/null || true)"
  RESOLUTION="$(jq -r '.resolution // .fix // .session.resolution // ""' "${TMP_INPUT}" 2>/dev/null || true)"
  NEXT_STEP="$(jq -r '.next // .nextStep // .session.next // ""' "${TMP_INPUT}" 2>/dev/null || true)"
fi

# 候補が空なら、生入力の断片をQuestionへ退避
RAW_SNIPPET="$(head -c 4000 "${TMP_INPUT}" | oneline)"

if [ -z "${HOOK_TIMESTAMP}" ]; then
  HOOK_TIMESTAMP="${TS}"
fi

if [ -z "${TOPIC}" ]; then
  TOPIC="(auto) ${EVENT_NAME} checkpoint"
fi

if [ -z "${QUESTION}" ]; then
  QUESTION="$(clip "${RAW_SNIPPET}" 400)"
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
  if [ "${EVENT_NAME}" = "PreCompact" ]; then
    NEXT_STEP="コンパクション後に未解決事項の確認と継続"
  elif [ "${EVENT_NAME}" = "Stop" ]; then
    NEXT_STEP="次回セッションで継続"
  else
    NEXT_STEP="未解決事項の確認と継続"
  fi
fi

if [ -z "${SESSION_ID}" ]; then
  SESSION_ID="(unknown)"
fi

if [ -z "${TRIGGER}" ]; then
  TRIGGER="(none)"
fi

TRANSCRIPT_SOURCE="(none)"
TRANSCRIPT_SNAPSHOT="(none)"
TRANSCRIPT_HASH="(none)"
TRANSCRIPT_BYTES="0"
TRANSCRIPT_STATUS="(not captured)"

if [ -n "${TRANSCRIPT_PATH}" ] && [ -r "${TRANSCRIPT_PATH}" ]; then
  TRANSCRIPT_SOURCE="${TRANSCRIPT_PATH}"
  TRANSCRIPT_HASH="$(hash_file "${TRANSCRIPT_PATH}")"
  TRANSCRIPT_BYTES="$(wc -c < "${TRANSCRIPT_PATH}" | awk '{print $1}')"

  EXISTING_SNAPSHOT=""
  if [ -f "${INDEX_FILE}" ]; then
    EXISTING_SNAPSHOT="$(awk -F '\t' -v hash="${TRANSCRIPT_HASH}" '$1==hash{print $2; exit}' "${INDEX_FILE}" 2>/dev/null || true)"
  fi

  if [ -n "${EXISTING_SNAPSHOT}" ] && [ -f "${EXISTING_SNAPSHOT}" ]; then
    TRANSCRIPT_SNAPSHOT="$(to_workspace_relative "${EXISTING_SNAPSHOT}")"
    TRANSCRIPT_STATUS="(reused existing snapshot)"
  else
    SAFE_EVENT="$(safe_token "${EVENT_NAME}")"
    SAFE_SESSION="$(safe_token "${SESSION_ID}")"
    SNAP_TS="$(date +%Y%m%dT%H%M%S%z | tr -d ':')"
    SNAPSHOT_PATH="${TRANSCRIPT_DIR}/${DD}_${SAFE_SESSION}_${SAFE_EVENT}_${SNAP_TS}.json"
    cp "${TRANSCRIPT_PATH}" "${SNAPSHOT_PATH}"
    printf '%s\t%s\n' "${TRANSCRIPT_HASH}" "${SNAPSHOT_PATH}" >> "${INDEX_FILE}"
    TRANSCRIPT_SNAPSHOT="$(to_workspace_relative "${SNAPSHOT_PATH}")"
    TRANSCRIPT_STATUS="(captured)"
  fi
fi

# Markdown追記
{
  echo ""
  echo "## JournalCheckpoint ${TS}"
  echo "- CheckpointId: ${CHECKPOINT_ID}"
  echo "- Event: ${EVENT_NAME}"
  echo "- HookTimestamp: ${HOOK_TIMESTAMP}"
  echo "- SessionId: ${SESSION_ID}"
  echo "- Trigger: ${TRIGGER}"
  echo "- Topic: ${TOPIC}"
  echo "- Question: ${QUESTION}"
  echo "- Blocker: ${BLOCKER}"
  echo "- Resolution: ${RESOLUTION}"
  echo "- Next: ${NEXT_STEP}"
  echo "- TranscriptPath: ${TRANSCRIPT_SOURCE}"
  echo "- TranscriptSnapshot: ${TRANSCRIPT_SNAPSHOT}"
  echo "- TranscriptHash: ${TRANSCRIPT_HASH}"
  echo "- TranscriptBytes: ${TRANSCRIPT_BYTES}"
  echo "- TranscriptStatus: ${TRANSCRIPT_STATUS}"
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