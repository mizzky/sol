# 学習記録 - 2026/02/02

## 📚 本日学習した技術パターンとその活用方法

### 1. テーブル駆動テスト（Table-Driven Tests）の活用

**学習内容**:
Go のテストにおいて、複数のテストケースを `[]struct` で定義し、ループで実行する手法を採用。

**実装パターン**:
```go
tests := []struct {
	name           string
	requestBody    map[string]interface{}
	setupMock      func(m *MockDB)
	setupTokenMock func(tg *MockTokenGenerator)
	expectedStatus int
	checkResponse  func(*testing.T, *httptest.ResponseRecorder)
}{
	{
		name: "正常系：ログイン成功",
		requestBody: map[string]interface{}{...},
		expectedStatus: http.StatusOK,
		setupMock: func(m *MockDB) {...},
		setupTokenMock: func(tg *MockTokenGenerator) {...},
		checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {...},
	},
	// 追加のテストケースを列挙
}

for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		// テストの実装
	})
}
```

**メリット**:
- テストケースの追加が容易（新しい struct を配列に追加するだけ）
- テスト結果の可読性が向上（各ケースの名前が表示される）
- DRY 原則に従い、テストコード内の重複を削減
- テストの保守性が向上

**いつ使うか**:
- 同じ関数を異なるパラメータで複数回テストする場合
- 正常系・異常系のテストケースが多い場合
- エッジケースを網羅的にテストしたい場合

---

### 2. モックの責務分離（Separation of Concerns）

**学習内容**:
テスト内でモック設定を行う際、各モック設定関数に明確な責務を割り当てることで、コードの可読性と保守性を向上させる。

**改善前（責務が混在）**:
```go
setupMock: func(m *MockDB) auth.TokenGenerator {
	// DB モック設定
	m.On("GetUserByEmail", ...).Return(...)
	
	// トークン生成モック設定
	mockTokenGenerator := new(MockTokenGenerator)
	mockTokenGenerator.On("GenerateToken", ...).Return(...)
	
	return mockTokenGenerator
}
```

**改善後（責務が分離）**:
```go
setupMock: func(m *MockDB) {
	// DB モック設定のみ
	m.On("GetUserByEmail", ...).Return(...)
}

setupTokenMock: func(tg *MockTokenGenerator) {
	// トークン生成モック設定のみ
	tg.On("GenerateToken", ...).Return(...)
}
```

**メリット**:
- 各関数の責務が明確になり、保守性が向上
- テストケースの意図が読み取りやすくなる
- 今後の拡張時に、新しいモック関数の追加が容易
- テストコードの可読性が大幅に向上

**いつ使うか**:
- 複数のモックオブジェクトを操作する場合
- テストの複雑さが増加している場合
- チーム開発で、コードレビューの効率を向上させたい場合

---

### 3. 依存性注入（Dependency Injection）によるテスト性の向上

**学習内容**:
外部の処理に依存する関数をテストする場合、その依存関係をインターフェース化し、テスト時にモック化することで、広範なテストケースが実現可能になる。

**改善前（直接呼び出し）**:
```go
func LoginUserHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ...
		token, err := auth.GenerateToken(user.ID) // 直接呼び出し
		if err != nil {
			// テストでエラーを再現できない
			RespondError(c, http.StatusInternalServerError, "...")
			return
		}
		// ...
	}
}
```

**改善後（インターフェース経由）**:
```go
func LoginUserHandler(q db.Querier, tokenGenerator auth.TokenGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ...
		token, err := tokenGenerator.GenerateToken(user.ID) // インターフェース経由
		if err != nil {
			// テストでエラーを再現可能
			RespondError(c, http.StatusInternalServerError, "...")
			return
		}
		// ...
	}
}
```

**テスト例**:
```go
mockTokenGenerator := new(MockTokenGenerator)
mockTokenGenerator.On("GenerateToken", int64(1)).
	Return("", errors.New("トークン生成エラー"))

router.POST("/api/login", handler.LoginUserHandler(mockDB, mockTokenGenerator))
```

**メリット**:
- トークン生成エラーなど、通常は発生しにくいエラーをテストできる
- テストの覆軄範囲が大幅に向上
- 実装の柔軟性が向上（トークン生成ロジックを差し替え可能）
- 本番環境では `DefaultTokenGenerator` を注入、テスト環境では `MockTokenGenerator` を注入

**いつ使うか**:
- 外部 API や時間に依存する処理をテストする場合
- 例外的なエラーをテストしたい場合
- モック化してテストしたい処理がある場合

---

### 4. モック設定のデフォルト動作

**学習内容**:
`testify/mock` を使用していて、「予期しないメソッド呼び出し」エラーが発生する場合、テストループ内でデフォルトの動作を設定することで回避可能。

**エラーが発生していた状況**:
```go
setupTokenMock: nil // モック設定がない
// LoginUserHandler が tokenGenerator.GenerateToken(user.ID) を呼び出す
// ↓
// panic: I don't know what to return because the method call was unexpected.
```

**解決策**:
```go
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		mockTokenGenerator := new(MockTokenGenerator)

		// デフォルト動作を設定
		if tt.setupTokenMock != nil {
			tt.setupTokenMock(mockTokenGenerator)
		} else {
			mockTokenGenerator.On("GenerateToken", mock.Anything).
				Return("default_token", nil)
		}

		// テストの実行
	})
}
```

**メリット**:
- すべてのテストケースで、トークン生成が成功することを前提にできる
- `setupTokenMock` を明示的に設定する必要があるのは、エラー動作をテストする場合のみ
- テストコードの簡潔性が向上
- テストケースの追加が容易

**いつ使うか**:
- 複数のテストケースで同じモックオブジェクトを使用する場合
- デフォルトの動作と例外的な動作を分け持たせたい場合

---

## 🎯 適用した設計パターンの関係図

```
テーブル駆動テスト
  ↓
各テストケースで setupMock / setupTokenMock を指定
  ↓
モックの責務分離
  ↓
各モック関数が特定の依存関係を設定
  ↓
依存性注入
  ↓
LoginUserHandler に依存関係を注入
```

---

## 💡 今後への応用例

### 1. `RegisterUserHandler` のテスト設計
- テーブル駆動テストで、重複メール、パスワード不正などの異常系をカバー
- `setupMock` で DB モックの挙動を定義
- `setupMockForPassword` を追加し、パスワードハッシュ化の動作を制御

### 2. 商品 CRUD API のテスト設計
- カテゴリと同様のパターンを適用
- より多くの依存関係（カテゴリ存在確認など）をモック化

### 3. トランザクションを伴うテスト
- 複数の DB 操作をモック化し、トランザクション内のエラーをテスト
- ロールバック動作を検証

---

## 📊 本日のテスト統計

- **テストケース数**：5（正常系1件、異常系4件）
- **テストカバレッジ**：`user.go` でほぼ100%（トークン生成エラーケース含む）
- **実行時間**：~150ms
- **成功率**：100%（すべてのテストが成功）

---

## 📝 ベストプラクティス

1. **テストケースの命名**
   - `正常系：ログイン成功` のように、テストの目的を明確に表現

2. **モック設定の明確性**
   - 各モック設定関数に明確な責務を割り当て
   - デフォルト動作を設定して予期しないエラーを防止

3. **依存性注入の活用**
   - 外部依存関係をインターフェース化
   - テスト時にモック化可能にする

4. **テストコードの DRY 性**
   - テーブル駆動テストで重複を削減
   - 共通のモック設定を関数に抽出

---

**作成日時**: 2026-02-02 15:50  
**ステータス**: ✅ 学習記録完了
