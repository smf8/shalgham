create table if not exists conversations
(
    id          serial,
    name     text        not null unique,
    primary key (id)
);