CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS payments (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id       UUID NOT NULL UNIQUE,
    user_id        UUID NOT NULL,
    amount         DECIMAL(12,2) NOT NULL,
    currency       VARCHAR(3) NOT NULL DEFAULT 'USD',
    status         VARCHAR(50) NOT NULL DEFAULT 'pending',
    method         VARCHAR(50) NOT NULL,
    transaction_id VARCHAR(255) NOT NULL,
    created_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);
