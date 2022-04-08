create table links
(
    id         serial,
    token      varchar(50) not null,
    url        text        not null,
    created_at timestamp   not null,
    clicks     int
);

create unique index links_token_uindex
    on links (token);

create unique index links_url_uindex
    on links (url);

