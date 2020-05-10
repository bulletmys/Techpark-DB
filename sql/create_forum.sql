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
