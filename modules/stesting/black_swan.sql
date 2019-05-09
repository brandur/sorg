BEGIN TRANSACTION;

CREATE EXTENSION IF NOT EXISTS hstore WITH SCHEMA public;

DROP TABLE IF EXISTS events;

CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    slug text NOT NULL,
    content text,
    occurred_at timestamp without time zone NOT NULL,
    metadata hstore,
    type text
);

COMMIT TRANSACTION;
