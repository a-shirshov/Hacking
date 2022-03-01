drop table if exists "request-response";

create table "request-response" (
    id serial not null UNIQUE,
    request TEXT not null,
    response TEXT not null,
    requestJson TEXT not null,
    responseJson TEXT not null,
    isSecure bool not null
);

