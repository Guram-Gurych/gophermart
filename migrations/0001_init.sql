CREATE TABLE users (
    id UUID PRIMARY KEY,
    login varchar(255) UNIQUE NOT NULL,
    password_hash varchar(255)
);

CREATE TABLE orders
(
    number VARCHAR(255) PRIMARY KEY,
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'NEW',
    accrual BIGINT DEFAULT 0,
    uploaded_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES Users(id),
    CONSTRAINT check_status CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED'))
);

CREATE TABLE withdrawal
(
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    order_number VARCHAR(255) UNIQUE NOT NULL,
    sum BIGINT NOT NULL DEFAULT 0,
    processed_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES Users (id)
);

CREATE TABLE balances
(
    user_id UUID PRIMARY KEY REFERENCES users(id),
    current BIGINT NOT NULL DEFAULT 0,
    withdrawn BIGINT NOT NULL DEFAULT 0
);
