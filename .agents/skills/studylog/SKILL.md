---
name: studylog
description: セッション中の学習記録内容を自動記録。疑問、躓き、解決策、次課題をキャプチャする。
---

# Study Log Skill

## 出力先
- 中間ジャーナル（自動）: `doc/memo/YYYY/MM/DD_journal.tmp.md`
- transcriptスナップショット（自動）: `doc/memo/YYYY/MM/transcripts/*.json`
- 学習ログ（最終）: `doc/memo/YYYY/MM/DD_learning.md`

## 処理構造

| フェーズ | トリガー | 担当 | 処理内容 |
| --- | --- | --- | --- |
| チェックポイント追記 | PreCompact / Stop（自動） | Shell | ジャーナルへチェックポイント追記 + transcript_path の実体を snapshot として永続化 |
| 直近プロンプト退避 | UserPromptSubmit（自動） | Shell | ログ記録直前を含むユーザープロンプト送信時に transcript snapshot を保存 |
| 最終化チェックポイント追記 | ログ記録指示 | Shell | FinalizeLog をジャーナルに追記 |
| 学習ログ生成 | ログ記録指示 | LLM | ジャーナル + 永続化済み snapshot を読んで要約・整形し learning.md に書き込む |

## ログ記録指示を受けたときの手順

### ステップ1: 最終チェックポイントを追記してジャーナルパスを取得する

```sh
cd /workspaces/sol_coffeesys
sh .agents/skills/studylog/scripts/finalize-studylog.sh "ログ記録"
```

スクリプトが stdout にジャーナルファイルのパスを出力する。

### ステップ2: ジャーナルを読み込む

ステップ1で得たパスのファイルを全文読み込む。

### ステップ2.1: ジャーナルに記録された transcript snapshot をすべて読み込む

- ジャーナル中の `- TranscriptSnapshot:` 行を抽出する。
- 値が `(none)` 以外のパスをユニーク化して、すべて読み込む。
- snapshot が多い場合でも省略せず、少なくとも全ファイルの見出し情報と末尾の更新部分を確認する。
- `Event: PreCompact` の snapshot は、コンパクション前の文脈を保全する一次情報として最優先で確認する。
- `Event: UserPromptSubmit` の snapshot は、ログ記録指示の直前までの会話を補完するために確認する。

### ステップ3: 学習ログを生成して learning.md に書き込む

ジャーナル + snapshot 内容をもとに、LLM が以下のルールで `doc/memo/YYYY/MM/DD_learning.md` を書き込む。
シェルスクリプトには頼らず、LLM 自身がファイル書き込みを行う。

#### ファイルが存在しない場合

```markdown
# 学習記録 YYYY-MM-DD

## セッション 1 (HH:MM)

### 取り組んだタスク
（ジャーナルから読み取ったタスク名・作業内容を記載）

### ユーザーが質問した内容
（セッション中にユーザーが質問したことを箇条書きで列挙）

### 躓いたポイントと解決策
（各躓きにつき原因・解決策をセットで記載）

### 次回課題
（未解決事項・継続タスク・次に取り組むべきことを列挙）
```

#### ファイルが既に存在する場合

既存の内容は一切変更せず、末尾に以下を追記する。
セッション番号は既存セクション数 + 1 にする。

```markdown

---

## セッション N (HH:MM)

### 取り組んだタスク
（ジャーナルから読み取ったタスク名・作業内容を記載）

### ユーザーが質問した内容
（セッション中にユーザーが質問したことを箇条書きで列挙）

### 躓いたポイントと解決策
（各躓きにつき原因・解決策をセットで記載）

### 次回課題
（未解決事項・継続タスク・次に取り組むべきことを列挙）
```
