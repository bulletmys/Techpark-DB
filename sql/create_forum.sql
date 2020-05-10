-- CREATE TABLE tracks
-- (
--     ID        BIGSERIAL PRIMARY KEY,
--     name      VARCHAR(100) NOT NULL,
--     duration  INTEGER      NOT NULL,
--     image     VARCHAR DEFAULT '/static/img/track/default.png',
--     link      VARCHAR      NOT NULL,
--     artist_id BIGSERIAL    NOT NULL,
--     FOREIGN KEY (artist_ID) REFERENCES artists (ID)
--         ON DELETE CASCADE
--         ON UPDATE CASCADE
-- );


CREATE TABLE users
(
    nick  VARCHAR PRIMARY KEY,
    email VARCHAR UNIQUE,
    name  VARCHAR NOT NULL,
    about VARCHAR
);

DROP TABLE if exists forums;
CREATE TABLE forums
(
    slug    VARCHAR PRIMARY KEY,
    nick    VARCHAR NOT NULL,
    title   VARCHAR NOT NULL,
    threads INT    DEFAULT 0,
    posts   BIGINT DEFAULT 0,
    FOREIGN KEY (nick) REFERENCES users (nick)
);

CREATE TABLE threads
(
    nick    VARCHAR NOT NULL,
    created timestamp,
    forum   VARCHAR NOT NULL,
    id      BIGSERIAL PRIMARY KEY,
    message VARCHAR NOT NULL,
    slug    VARCHAR UNIQUE,
    title   VARCHAR NOT NULL,
    votes   int default 0,
    FOREIGN KEY (nick) REFERENCES users (nick),
    FOREIGN KEY (forum) REFERENCES forums (slug)
)
