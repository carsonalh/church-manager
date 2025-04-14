CREATE TABLE member (
    id BIGSERIAL NOT NULL,
    first_name VARCHAR(128),
    last_name VARCHAR(128),
    email_address VARCHAR(256),
    phone_number VARCHAR(128),
    -- can be the empty string if unused
    notes TEXT NOT NULL
);