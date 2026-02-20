# 2026-02-20 学習記録

## 本日の学習テーマ
TDD（テスト駆動開発）を用いたフロントエンド開発の実践

---

## 実施したタスク

### チケット11: トップページと管理ページの分離（完了）
- User インターフェースへの `role` フィールド追加
- Header コンポーネントの TDD 実装
- AdminRoute コンポーネントの基本実装
- トップページのリファクタリング（商品登録フォーム削除）
- `/admin/products` 管理ページの新規作成
- RootLayout への Header 適用

### チケット12: 管理者権限チェックミドルウェア強化（完了）
- AdminRoute の非同期処理対応（loadFromStorage 完了待ち）
- 401 エラー時の自動ログアウト処理
- 既存テストの非同期対応修正

---

## 学んだこと

### 1. TDD サイクルの実践
**Red → Green → Refactor のサイクル**
- **Red**: まず失敗するテストを書く
  - 例：`AdminRoute.behavior.test.tsx` で「ロード中はリダイレクトしない」というテストを先に作成
- **Green**: 最小限の実装でテストをパスさせる
  - 例：`restoring` state を追加して `loadFromStorage()` 完了を待つロジックを実装
- **Refactor**: コードを整理（今回は最小実装で完了したため省略）

**気づき**: テストを先に書くことで「何を実装すべきか」が明確になり、実装の方針がブレなかった。

---

### 2. React Testing Library での非同期処理のテスト

#### 問題: act 警告が頻発
```
console.error
  An update to AdminRoute inside a test was not wrapped in act(...).
```

#### 原因
- コンポーネント内の非同期 state 更新（`setRestoring(false)`）がテストの予期しないタイミングで発生
- `loadFromStorage()` が Promise を返すが、テスト側で完了を待っていなかった

#### 解決方法
1. **render を act でラップ**
   ```typescript
   await act(async () => {
     render(<AdminRoute><TestComponent /></AdminRoute>);
   });
   ```

2. **loadFromStorage をモックで async 関数に**
   ```typescript
   useAuthStore.setState({
     loadFromStorage: jest.fn(async () => {}),
     // ...
   } as any);
   ```

3. **アサーションを waitFor でラップ**
   ```typescript
   await waitFor(() => {
     expect(mockRouter.push).toHaveBeenCalledWith("/login");
   });
   ```

**学び**: React の非同期更新をテストする際は `act` と `waitFor` を適切に使い分ける必要がある。

---

### 3. Next.js の useRouter モック

#### 問題: router.replace が undefined
```
TypeError: router.replace is not a function
```

#### 原因
- テストで `mockRouter = { push: jest.fn() }` としてモックしていたが、実装で `router.replace()` を使っていた

#### 解決
- 実装を `router.push()` に統一（Next.js の App Router では push と replace の挙動はほぼ同じ）
- または、モックに `replace` も追加する選択肢もあった

**学び**: モックと実装の整合性を常に確認する。テストが失敗した際は「モックが不足していないか」を疑う。

---

### 4. Zustand の state モック

#### 課題
- テスト中に `useAuthStore` の state を動的に変更する必要があった
- 各テストケースで異なる認証状態（未ログイン、member、admin）をシミュレート

#### 解決パターン
```typescript
beforeEach(() => {
  useAuthStore.setState({
    token: null,
    user: null,
    loadFromStorage: jest.fn(async () => {}),
    logout: jest.fn(),
  } as any);
});

it("admin の場合", async () => {
  useAuthStore.setState({
    token: "t",
    user: { id: 2, name: "Admin", email: "a@e", role: "admin" },
    loadFromStorage: jest.fn(async () => {}),
    logout: jest.fn(),
  } as any);
  // ...
});
```

**学び**: Zustand の `setState` を使えばテスト内で簡単に state を操作できる。`as any` で型エラーを回避しつつ、必要な関数だけモック化。

---

### 5. TypeScript 型安全性の重要性

#### 問題
- バックエンドの User モデルには `role` フィールドがあるが、フロントエンドの User インターフェースには未定義だった
- テストコードでは `role: "member"` を参照していたため、型不整合が発生

#### 解決
```typescript
// frontend/store/useAuthStore.ts
export interface User {
  id: number;
  name: string;
  email: string;
  role: "admin" | "member";  // ← 追加
}
```

**学び**: フロントエンドとバックエンドの型定義を一致させることで、実行時エラーを防げる。TypeScript のユニオン型（`"admin" | "member"`）で不正な値を防止。

---

## 詰まったポイントと解決策

### 1. テストが非同期処理を待たずに失敗する
**症状**: 
- `expect(mockRouter.push).toHaveBeenCalledWith("/login")` が `Number of calls: 0` で失敗
- コンポーネントは正しく動作しているのにテストだけ失敗

**原因**: 
- `loadFromStorage()` が非同期だが、テストが同期的にアサーションを実行していた

**解決**: 
- `await act(async () => { render(...) })` でレンダリングを待つ
- `await waitFor(() => { expect(...) })` でアサーションを待つ

**教訓**: React の非同期処理は「完了を待つ」ことが必須。同期的なテストは必ず失敗する。

---

### 2. 複数のテストファイルで同じエラーが出る
**症状**: 
- `AdminRoute.test.tsx` と `AdminPage.test.tsx` で同じ act 警告が出る
- 新しく作った `AdminRoute.behavior.test.tsx` だけパスする

**原因**: 
- 古いテストファイルが非同期処理に対応していなかった

**解決**: 
- 既存テストファイル2つを同じパターンで修正（`act` + `waitFor` 追加）

**教訓**: 実装を変更したら、関連する全てのテストを確認・修正する必要がある。

---

### 3. Header コンポーネントで「Admin」が2つマッチする
**症状**: 
```
TestingLibraryElementError: Found multiple elements with the text: /Admin/i
- <a href="/admin/products">Admin</a>
- <span>Admin User</span>
```

**原因**: 
- `screen.getByText(/Admin/i)` が Admin リンクとユーザー名（"Admin User"）の両方にマッチ

**解決**: 
```typescript
// Before
expect(screen.getByText(/Admin/i)).toBeInTheDocument();

// After
expect(screen.getByRole("link", { name: "Admin" })).toBeInTheDocument();
```

**教訓**: `getByText` よりも `getByRole` の方が要素を特定しやすい。アクセシビリティ的にも推奨される。

---

## 技術的な学び

### React Testing Library のベストプラクティス
1. **ユーザー視点でテストを書く**: DOM 構造ではなく、ユーザーが見る要素（role, text）でセレクト
2. **非同期処理は必ず待つ**: `act`, `waitFor`, `findBy...` を適切に使い分ける
3. **テストは独立させる**: `beforeEach` で state をリセットし、テスト間の依存を排除

### TDD のメリット（実感）
- **設計が明確になる**: テストを書くことで「何が必要か」が整理される
- **リファクタリングが安全**: テストがあるので、コード変更後も動作を保証できる
- **デバッグが楽**: テストが失敗した時点で問題箇所が分かる

### TDD で苦労したこと
- **テストの書き方が分からない**: 最初は「何をテストすべきか」が不明確だった
  - 解決: ユーザーストーリー（「未ログインなら /login に飛ぶ」）をテストケースに変換
- **モックの設定が複雑**: useRouter, useAuthStore など、外部依存が多い
  - 解決: 各テストファイルの `beforeEach` でモックを統一パターンで設定

---

## 次回への課題

### 1. サーバーサイドの保護
- 現在はクライアント側（AdminRoute）でのガードのみ
- Next.js の middleware やサーバーコンポーネントでの認証チェックも検討すべき

### 2. E2E テストの追加
- 単体テストは完璧だが、実際のブラウザでの動作確認はまだ
- Playwright や Cypress での自動テストを検討

### 3. テストカバレッジの可視化
- `jest --coverage` を実行してカバレッジを確認
- 80% 以上を目標にする

---

## 本日の成果

### 実装
- チケット11（4サブタスク）完了
- チケット12 完了
- 全テスト: **37 件パス、0 件失敗**

### Git コミット
```bash
# チケット11
git commit -m "feat(header): add Header navigation component with authentication state"
git commit -m "feat(page): separate admin product form into /admin/products and add Header to layout (TDD)"

# チケット12
git commit -m "feat(admin): wait loadFromStorage in AdminRoute and handle logout on invalid token"
```

### コード変更量
- 新規ファイル: 5 個
  - `Header.tsx`, `Header.test.tsx`
  - `AdminRoute.tsx`, `AdminRoute.test.tsx`, `AdminRoute.behavior.test.tsx`
  - `admin/products/page.tsx`
- 修正ファイル: 4 個
  - `useAuthStore.ts` (User interface に role 追加)
  - `api.ts` (LoginResponse に role 追加)
  - `page.tsx` (フォーム削除)
  - `layout.tsx` (Header 適用)

---

## 感想

TDD は最初は手間に感じたが、慣れてくると「失敗するテストを書く→実装する→パスする」というリズムが心地よかった。特に、実装後にテストが全てグリーンになる瞬間は達成感がある。

非同期処理のテストは難しかったが、`act` と `waitFor` のパターンを理解できたのが大きな収穫。今後は自信を持って非同期コンポーネントのテストを書ける。

型安全性の重要性を再認識した。バックエンドとフロントエンドの型を一致させることで、実行時エラーを防げる。

---

## 参考資料

- [React Testing Library 公式ドキュメント](https://testing-library.com/docs/react-testing-library/intro/)
- [Testing Library: Common mistakes](https://kentcdodds.com/blog/common-mistakes-with-react-testing-library)
- [Next.js useRouter hook](https://nextjs.org/docs/app/api-reference/functions/use-router)
- [Zustand Testing](https://docs.pmnd.rs/zustand/guides/testing)

---

作成日: 2026-02-20
