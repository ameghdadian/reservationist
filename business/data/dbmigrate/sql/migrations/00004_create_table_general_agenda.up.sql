CREATE TABLE IF NOT EXISTS general_agenda (
    id              UUID        NOT NULL,
    business_id     UUID        UNIQUE NOT NULL,
    opens_at        TIMESTAMP   NOT NULL,
    closed_at       TIMESTAMP   NOT NULL,
    interval        INTEGER     NOT NULL CHECK(interval > 0 AND interval <= 86400),
    working_days    INTEGER[]   NOT NULL,
    date_created    TIMESTAMP   NOT NULL,
    date_updated    TIMESTAMP   NOT NULL,

    PRIMARY KEY(id),
    FOREIGN KEY (business_id) REFERENCES businesses(business_id) ON DELETE CASCADE
);