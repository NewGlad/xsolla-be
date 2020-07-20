CREATE TABLE users (
    id bigserial not null primary key,
    username varchar not null unique,
    encrypted_password varchar not null
);