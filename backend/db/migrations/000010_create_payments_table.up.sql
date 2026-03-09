CREATE TABLE IF NOT EXISTS payments(
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
    amount BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, completed, failed
    payment_method VARCHAR(50),
    external_transaction_id VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);