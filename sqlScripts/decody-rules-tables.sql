create table if not exists files
(
    file_name varchar(128) not null,
    name      varchar(128) not null,
    file_id   int auto_increment
        primary key,
    constraint files_uk
        unique (file_name, name)
);

create table if not exists rules
(
    rule_id     int auto_increment
        primary key,
    file_id     int           not null,
    `condition` varchar(2048) not null,
    explanation varchar(2048) null,
    constraint rules___fk
        foreign key (file_id) references files (file_id)
);

