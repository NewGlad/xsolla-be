CREATE TABLE news (
    id bigserial not null primary key,
    content text not null,
    author_id bigint REFERENCES users(id),
    likes bigint not null
)
