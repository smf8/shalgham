create table if not exists users
(
    id        serial,
    username  varchar(255) not null unique,
    password  varchar(255) not null,
    is_online boolean,
    primary key (id)
);