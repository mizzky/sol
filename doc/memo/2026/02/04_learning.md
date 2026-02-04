# 2026年2月4日 学習記録

## 学習内容

### nullと空文字のAPI設計に関する学び
- **null**: 値が存在しないことを明示的に示す。
- **空文字**: 値が存在するが内容が空であることを示す。
- リクエストとレスポンスでの扱いを明確にし、一貫性を保つことが重要。
- 外部システムやフロントエンドとの連携を考慮した設計が必要。

### テスト駆動開発（TDD）の実践
- テストコードを先に書き、失敗する状態からプロダクトコードを改良。
- テストが失敗した場合、原因を特定し、修正を行う。
- `go test -coverprofile=coverage.out ./...` コマンドを使用してカバレッジを確認。

### デバッグの経験
- ハンドラー関数のリネームによるテスト失敗を解決。
- `sqlc`で生成された関数名とモックの整合性を保つ重要性を学んだ。

## 今後の課題
- APIドキュメントに`null`と空文字の扱いに関するルールを追記。
- バリデーションロジックのテストを強化。
- 外部システムとのデータ連携テストを計画。

## その他の学び
### エディタの分割
- 同じファイルの中で上の方と下の方見たかったけどファイル名とかが邪魔だった
- -> 表示->エディターレイアウト->グループ内分割 ```Ctrl+k, Ctrl+Shift+\```
---

### エディタの設定
```
{
  // --- 見た目のカスタマイズ ---
  "workbench.colorTheme": "Tokyo Night", // インストール済みのテーマ名
  "editor.fontFamily": "'Cascadia Code', Consolas, 'Courier New', monospace",
  "editor.fontLigatures": true, // 合字（-> が矢印になる）を有効化
  "window.titleBarStyle": "custom",

  // --- タブ・UIの非表示設定（ミニマル化） ---
  "workbench.editor.showTabs": "none", // タブ（ファイル名）を非表示
  "editor.minimap.enabled": false, // ミニマップ自体不要ならfalse
  "workbench.activityBar.location": "hidden", // 左側のアイコンバーも消して究極にスッキリさせる場合
  "workbench.statusBar.visible": true, // 下のバーは情報の宝庫なので残すのがおすすめ

  // --- コーディング補助 ---
  "editor.stickyScroll.enabled": true, // スティックスクロール
  "editor.breadcrumbs.enabled": true, // タブの代わりのパンくずリスト
  "editor.guides.bracketPairs": "active", // 括弧に色付け
  "editor.formatOnSave": true, // 保存時に自動でコードを綺麗にする

  // --- Markdown設定 ---
  "editor.wordWrap": "on" // 長い行を折り返す（Markdownが読みやすくなる）
}
```

本日の学びを活かし、引き続きAPI設計とテストの改善に取り組む。