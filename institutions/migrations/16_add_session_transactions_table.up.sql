CREATE TABLE
    session_transactions (
        id BIGSERIAL,
        session BIGINT NOT NULL,
        transaction BIGINT NOT NULL,
        description TEXT,
        key TEXT NOT NULL CHECK (key <> ''),
        UNIQUE (key, transaction),
        PRIMARY KEY (id),
        FOREIGN KEY (session) REFERENCES enrollment_sessions (id)
    );