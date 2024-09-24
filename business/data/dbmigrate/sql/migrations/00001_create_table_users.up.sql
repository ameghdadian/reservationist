CREATE TABLE IF NOT EXISTS users (
    user_id UUID NOT NULL,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    phone_no TEXT UNIQUE NOT NULL,
    roles TEXT[] NOT NULL,
    password_hash TEXT NOT NULL,
    enabled BOOLEAN NOT NULL,
    date_created TIMESTAMP NOT NULL,
    date_updated TIMESTAMP NOT NULL,
    
    PRIMARY KEY (user_id)
);