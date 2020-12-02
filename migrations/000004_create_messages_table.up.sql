create table if not exists messages
(
    id              serial,
    conversation_id int       not null,
    from_id         int       not null,
    body            text,
    created_at      timestamp not null default now(),
    primary key (id),
    FOREIGN KEY (conversation_id) REFERENCES conversations (id),
    FOREIGN KEY (from_id) REFERENCES users (id)
);