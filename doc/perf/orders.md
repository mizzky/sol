# ListOrderByUser

```sql
SELECT
    id, user_id, total, status, created_at, updated_at
FROM orders
WHERE user_id = :user_id
ORDER BY created_at DESC
LIMIT 50
OFFSET :offset;
```

## before 計測

- `backend/perf/sql/queries/orders_by_user_offset.sql`を使用する
- `:offset`を`0 / 1,000 / 10,000 / 100,000`と切り替えて計測する
- それぞれ3回ずつ計測し、中央値を測定する