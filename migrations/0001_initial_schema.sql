-- +migrate Up
CREATE TABLE accounts
(
    account_number VARCHAR(100) PRIMARY KEY,
    balance        INT NOT NULL
);

INSERT INTO accounts (account_number, balance)
VALUES ('schet_klienta', 2000000),
       ('schet_prosent_kredita', 0),
       ('schet_osnovnoy_dolg', 0),
         ('schet_penya_kredita', 0);

CREATE TABLE transactions
(
    id               SERIAL PRIMARY KEY,
    transaction_id   UUID        NOT NULL,
    transaction_type VARCHAR(100) NOT NULL,               -- e.g., 'day_open', 'day_close', etc.
    created_at       TIMESTAMP            DEFAULT NOW(),
    status           VARCHAR(100) NOT NULL DEFAULT 'NEW', -- 'NEW', 'SUCCESS', 'FAILED'
    message          TEXT,
    UNIQUE (transaction_id)
);

CREATE TABLE operations
(
    id             SERIAL PRIMARY KEY,
    operation_id   UUID        NOT NULL,
    amount         INT         NOT NULL,
    account_number VARCHAR(100) REFERENCES accounts (account_number),
    transaction_id UUID REFERENCES transactions (transaction_id) ON DELETE CASCADE,
    operation_type VARCHAR(100) NOT NULL,               -- 'credit', 'debit', 'refund'
    status         VARCHAR(100) NOT NULL DEFAULT 'NEW', -- 'NEW', 'SUCCESS', 'FAILED'
    created_at     TIMESTAMP            DEFAULT NOW(),
    UNIQUE (operation_id)
);

-- +migrate Down
DROP TABLE operations;
DROP TABLE transactions;
DROP TABLE accounts;
