CREATE TABLE IF NOT EXISTS businesses (
    business_id     UUID        NOT NULL,
    owner_id        UUID        NOT NULL,
    name            TEXT        NOT NULL,
    description     TEXT        NOT NULL,
    date_created    TIMESTAMP   NOT NULL,
    date_updated    TIMESTAMP   NOT NULL,

    PRIMARY KEY (business_id),
    FOREIGN KEY (owner_id) REFERENCES users(user_id) ON DELETE CASCADE
);