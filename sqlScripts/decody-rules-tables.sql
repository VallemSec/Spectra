create table if not exists files
(
    id        int auto_increment
        primary key,
    file_name varchar(1024) not null
);

create table if not exists key_value
(
    `key` varchar(1024) null,
    value text          not null,
    constraint `key`
        unique (`key`) using hash
);

create table if not exists rules
(
    id          int auto_increment
        primary key,
    category    varchar(256) default 'Server' not null,
    explanation text                          not null,
    `condition` text                          not null,
    file_id     int                           not null,
    name        text                          not null,
    constraint rules_ibfk_1
        foreign key (file_id) references files (id)
);

create index if not exists file_id
    on rules (file_id);
