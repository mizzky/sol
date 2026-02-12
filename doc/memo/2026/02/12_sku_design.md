# SKU/Barcode API Design Memo

## 1. HTTP Status Code
- **409 Conflict**
    - 重複登録（一意制約違反）が発生した際に使用。
    - 「サーバーの現在の状態と矛盾している」ことを示す。
    - ※構文エラー（400）やバリデーションエラー（422）と使い分ける。

## 2. ID Design (Hybrid Approach)
システム内部の結合用IDと、業務上の識別子を完全に切り離す。

- **Internal ID (Primary Key)**: `UUID` / `ULID`
    - サーバー側で自動生成。
    - 不変。DBのテーブル間リレーションに使用。
- **Business ID (Natural Key)**: `SKU` / `Barcode`
    - クライアントからリクエストに含めて送信。
    - 一意制約。検索や人間向けの識別に使用。
    - **主キーにしない理由**: 将来的なコード体系の変更による影響を最小限にするため。

## 3. Workflow
1. **Request (POST)**: `{"sku": "ABC-123", "name": "Product Name"}`
2. **Server Side**:
    - UUIDの新規発行。
    - `sku` の重複確認。
    - 既に存在すれば `409 Conflict` をレスポンス。
3. **Response (Success)**: 
    - `201 Created` と共に、発行された `id` (UUID) と `sku` を返却。

## 4. Key Benefits
- **堅牢性**: メーカー都合などでバーコードが変更されても、システム内のリレーション（UUID）を維持したまま値を更新できる。
- **利便性**: 実物のバーコードスキャナからの入力値（SKU）をそのまま業務識別子として扱える。
- **拡張性**: 将来的なマルチテナント化やデータ移行にも耐えうる設計。