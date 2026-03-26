---
name: studylog
description: "セッション中の学習記録内容を自動記録。疑問、躓き、解決策、次課題をキャプチャする"
---

# Study Log Skill

## 出力先
- 中間ジャーナル（自動）: `doc/memo/YYYY/MM/DD_journal.tmp.md`
- 学習ログ（最終）: `doc/memo/YYYY/MM/DD_learning.md`

## 処理構造

| フェーズ | トリガー | 担当 | 処理内容 |
|---------|---------|------|--------|
| チェックポイント追記 | PreCompact（自動）| Shell | ジャーナルに生情報を追記 |
| 最終化チェックポイント追記 | ログ記録指示 | Shell | FinalizeLog をジャーナルに追記 |
| 学習ログ生成 | ログ記録指示 | **LLM** | ジャーナルを読んで要約・整形し learning.md に書き込む |

## ログ記録指示を受けたときのLLM手順（必ず順番どおり実行すること）

### ステップ1: 最終チェックポイントを追記してジャーナルパスを取得する

```sh
cd /workspaces/sol_coffeesys
chmod +x .github/skills/studylog/scripts/auto-checkpoint.sh .github/skills/studylog/scripts/finalize-studylog.sh
./.github/skills/studylog/scripts/finalize-studylog.sh "ログ記録"
```

スクリプトが stdout にジャーナルファイルのパスを出力する。

### ステップ2: ジャーナルを読み込む

ステップ1で得たパスのファイルを全文読み込む。

### ステップ3: 学習ログを生成して learning.md に書き込む

ジャーナル内容をもとに、**LLMが以下のルールで** `doc/memo/YYYY/MM/DD_learning.md` を書き込む。
シェルスクリプトには頼らず、LLM自身がファイル書き込みを行うこと。

#### ファイルが存在しない場合 → 新規作成

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

#### ファイルが既に存在する場合 → 末尾に追記（上書き禁止）

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

## 中間ジャーナルのチェックポイント形式（参考）

```markdown
## JournalCheckpoint [時刻]
- CheckpointId: cp-...
- Event: PreCompact | FinalizeLog
- Topic: ...
- Question: ...
- Blocker: ...
- Resolution: ...
- Next: ...

### RawHookInputSnippet
...
```