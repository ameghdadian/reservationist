CREATE TABLE IF NOT EXISTS daily_agenda (
    id                  UUID        NOT NULL,
    business_id         UUID        NOT NULL,
    opens_at            TIMESTAMP   NULL,
    closed_at           TIMESTAMP   NULL,
    interval            INTEGER     NULL    CHECK(interval > 0 AND interval <= 86400),
    availability        BOOLEAN     NOT NULL,
    date_created        TIMESTAMP   NOT NULL,
    date_updated        TIMESTAMP   NOT NULL,

    PRIMARY KEY(id),
    FOREIGN KEY (business_id) REFERENCES businesses(business_id) ON DELETE CASCADE
);