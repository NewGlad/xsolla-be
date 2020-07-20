CREATE TABLE likes (
    user_id bigint REFERENCES users(id),
    news_id bigint REFERENCES news(id),
    CONSTRAINT unq_user_news UNIQUE(user_id, news_id)
);