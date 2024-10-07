CREATE TABLE IF NOT EXISTS appointments (
    appointment_id  UUID        NOT NULL,
    business_id     UUID        NOT NULL,
    user_id         UUID        NOT NULL,
    status          SMALLINT    NOT NULL,
    scheduled_on    TIMESTAMP   NOT NULL,
    date_created    TIMESTAMP   NOT NULL,
    date_updated    TIMESTAMP   NOT NULL,

    UNIQUE (user_id, scheduled_on),
    UNIQUE (business_id, scheduled_on),

    PRIMARY KEY(appointment_id),
    FOREIGN KEY (business_id) REFERENCES businesses(business_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);