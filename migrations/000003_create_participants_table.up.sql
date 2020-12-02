create table if not exists participants
(
    id              serial,
    conversation_id int       not null,
    user_id         int       not null,
    primary key (id),
    FOREIGN KEY (conversation_id) REFERENCES conversations(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);