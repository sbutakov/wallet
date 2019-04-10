DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_direction') THEN
        CREATE TYPE payment_direction AS ENUM ('outgoing', 'incoming');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS accounts (
    id         UUID        NOT NULL PRIMARY KEY,
    name       VARCHAR(50) NOT NULL,
    currency   VARCHAR(3)  NOT NULL,
    balance    NUMERIC     NOT NULL,
    created_at TIMESTAMP   WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payments (
    id         UUID              NOT NULL PRIMARY KEY,
    account    UUID              REFERENCES accounts(id),
    account_to UUID              REFERENCES accounts(id),
    amount     NUMERIC           NOT NULL,
    direction  payment_direction NOT NULL,
    created_at TIMESTAMP        WITH TIME ZONE NOT NULL DEFAULT NOW()
);
