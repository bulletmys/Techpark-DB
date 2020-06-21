create extension if not exists citext;
CREATE UNLOGGED TABLE users
(
    nick  CITEXT /*COLLATE ucs_basic*/ PRIMARY KEY,
    email CITEXT UNIQUE,
    name  CITEXT NOT NULL,
    about VARCHAR
);

CREATE UNLOGGED TABLE forums
(
    slug    CITEXT PRIMARY KEY,
    nick    CITEXT  NOT NULL,
    title   VARCHAR NOT NULL,
    threads INT    DEFAULT 0,
    posts   BIGINT DEFAULT 0,
    FOREIGN KEY (nick) REFERENCES users (nick)
);

CREATE UNLOGGED TABLE threads
(
    nick    CITEXT  NOT NULL,
    created timestamptz(6),
    forum   CITEXT  NOT NULL,
    id      BIGSERIAL PRIMARY KEY,
    message VARCHAR NOT NULL,
    slug    CITEXT UNIQUE,
    title   VARCHAR NOT NULL,
    votes   int default 0,
    FOREIGN KEY (nick) REFERENCES users (nick),
    FOREIGN KEY (forum) REFERENCES forums (slug)
);

CREATE OR REPLACE FUNCTION after_thread_insert_func() RETURNS TRIGGER AS
$after_thread_insert$
BEGIN
    update forums set threads = threads + 1 where NEW.forum = slug;
    RETURN NEW;
END;
$after_thread_insert$ LANGUAGE plpgsql;

CREATE TRIGGER after_thread_insert
    AFTER INSERT
    ON threads
    FOR EACH ROW
EXECUTE PROCEDURE after_thread_insert_func();

CREATE UNLOGGED TABLE posts
(
    nick     CITEXT  NOT NULL,
    created  timestamptz(6),
    forum    CITEXT  NOT NULL,
    id       BIGSERIAL PRIMARY KEY,
    isEdited bool    NOT NULL,
    message  VARCHAR NOT NULL,
    parent   BIGINT DEFAULT 0,
    thread   int     not null,
    path     BIGINT[],
    FOREIGN KEY (nick) REFERENCES users (nick),
    FOREIGN KEY (forum) REFERENCES forums (slug),
    FOREIGN KEY (thread) REFERENCES threads (id)
);

CREATE OR REPLACE FUNCTION after_post_insert_func() RETURNS TRIGGER AS
$after_post_insert$
DECLARE
    parentArray BIGINT[];
BEGIN
    update forums set posts = posts + 1 where NEW.forum = slug;
    IF (NEW.parent <> 0) THEN
        parentArray := (select path from posts where id = NEW.parent);
        NEW.path = parentArray || currval('posts_id_seq');
    else
        NEW.path = ARRAY [currval('posts_id_seq')];
    end if;
    RETURN NEW;
END;
$after_post_insert$ LANGUAGE plpgsql;

CREATE TRIGGER after_post_insert
    BEFORE INSERT
    ON posts
    FOR EACH ROW
EXECUTE PROCEDURE after_post_insert_func();

CREATE UNLOGGED TABLE votes
(
    nick   CITEXT  not null,
    vote   integer not null,
    thread int4    not null,
    FOREIGN KEY (nick) REFERENCES users (nick),
    FOREIGN KEY (thread) REFERENCES threads (id),
    CONSTRAINT vote_pk PRIMARY KEY (nick, thread)
);
--
-- CREATE OR REPLACE FUNCTION check_global_var(v bigint, cur bigint) RETURNS bool as
-- $$
-- DECLARE
--     p int4;
-- BEGIN
--     p := (SELECT current_setting('myapp.var1'));
--     if (cur > v and p = '1') then
--         SELECT set_config('myapp.var1', '0', false);
--         return (select true);
--     else
--         return (select false);
--     end if;
-- END;
-- $$ LANGUAGE plpgsql;
--
--
-- CREATE OR REPLACE FUNCTION get_all_foo(since int, threadID bigint) RETURNS int AS
-- $BODY$
-- DECLARE
--     r   posts%rowtype;
--     kek int;
-- BEGIN
--     kek := 0;
--     FOR r IN
--         SELECT * FROM posts WHERE thread = threadID order by path, id
--         LOOP
--             kek = kek + 1;
--             if r.id = since then
--                 RETURN kek;
--             end if;
--         END LOOP;
--     RAISE NOTICE '%', kek;
--     RETURN kek;
-- END
-- $BODY$
--     LANGUAGE plpgsql;
--
--
-- CREATE OR REPLACE FUNCTION get_parent_tree(limiting int, threadID bigint, dataOffset int) RETURNS SETOF posts AS
-- $BODY$
-- DECLARE
--     r   posts%rowtype;
--     kek int;
-- BEGIN
--     kek := 0;
--     FOR r IN
--         SELECT * FROM posts WHERE thread = threadID order by path, id offset dataOffset
--         LOOP
--             if r.parent = 0 then
--                 kek = kek + 1;
--             end if;
--             if kek > limiting then
--                 return;
--             end if;
--             RETURN next r;
--         END LOOP;
--     RETURN;
-- END
-- $BODY$
--     LANGUAGE plpgsql;
--
--
-- CREATE OR REPLACE FUNCTION lol() RETURNS posts AS
-- $BODY$
-- DECLARE
--     r   posts%rowtype;
--     r2  RECORD;
--     kek int;
-- BEGIN
--     kek := 0;
--     FOR r IN
--         SELECT * FROM posts WHERE parent = 0 order by id desc
--         LOOP
--             RETURN (select * from posts where parent = r.parent order by id);
--         END LOOP;
-- END
-- $BODY$
--     LANGUAGE plpgsql;
--
-- CREATE OR REPLACE FUNCTION get_all_foo2(since int, threadID bigint) RETURNS int AS
-- $BODY$
-- DECLARE
--     r   posts%rowtype;
--     kek int;
-- BEGIN
--     kek := 0;
--     FOR r IN
--         SELECT * FROM posts WHERE thread = threadID order by path desc, id
--         LOOP
--             kek = kek + 1;
--             if r.id = since then
--                 RETURN kek;
--             end if;
--         END LOOP;
--     RAISE NOTICE '%', kek;
--     RETURN kek;
-- END
-- $BODY$
--     LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION insert_vote() RETURNS TRIGGER AS
$insert_vote$
BEGIN
    UPDATE threads
    SET votes = votes + NEW.vote
    WHERE id = NEW.thread;
    RETURN NEW;
END;
$insert_vote$
    LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS insert_vote ON votes;
CREATE TRIGGER insert_vote
    BEFORE INSERT
    ON votes
    FOR EACH ROW
EXECUTE PROCEDURE insert_vote();


CREATE OR REPLACE FUNCTION update_vote() RETURNS TRIGGER AS
$update_vote$
BEGIN
    UPDATE threads
    SET votes = votes - OLD.vote + NEW.vote
    WHERE id = NEW.thread;
    RETURN NEW;
END;
$update_vote$
    LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS update_vote ON votes;
CREATE TRIGGER update_vote
    BEFORE UPDATE
    ON votes
    FOR EACH ROW
EXECUTE PROCEDURE update_vote();

CREATE UNLOGGED TABLE forum_users
(
    forum_user CITEXT COLLATE ucs_basic NOT NULL,
    forum      CITEXT                   NOT NULL,
    email      TEXT                     NOT NULL,
    fullname   TEXT                     NOT NULL,
    about      TEXT
);

CREATE OR REPLACE FUNCTION add_forum_user() RETURNS TRIGGER AS
$add_forum_user$
BEGIN
    INSERT INTO forum_users (forum_user, forum, email, fullname, about)
    SELECT nick, NEW.forum, email, name, about
    FROM users
    WHERE nick = NEW.nick
    ON CONFLICT DO NOTHING;
    RETURN NULL;
END;
$add_forum_user$
    LANGUAGE plpgsql;


CREATE TRIGGER add_forum_user
    AFTER INSERT
    ON threads
    FOR EACH ROW
EXECUTE PROCEDURE add_forum_user();

CREATE OR REPLACE FUNCTION add_forum_user_p() RETURNS TRIGGER AS
$add_forum_user_p$
BEGIN
    INSERT INTO forum_users (forum_user, forum, email, fullname, about)
    SELECT nick, NEW.forum, email, name, about
    FROM users
    WHERE nick = NEW.nick
    ON CONFLICT DO NOTHING;
    RETURN NULL;
END;
$add_forum_user_p$
    LANGUAGE plpgsql;


CREATE TRIGGER add_forum_user_p
    AFTER INSERT
    ON posts
    FOR EACH ROW
EXECUTE PROCEDURE add_forum_user_p();


CREATE UNIQUE INDEX IF NOT EXISTS idx_fu_user ON forum_users (forum, forum_user);
CREATE UNIQUE INDEX IF NOT EXISTS idx_forum_slug  ON forums (slug);

CREATE INDEX IF NOT EXISTS idx_threads_slug ON threads (slug);
CREATE INDEX IF NOT EXISTS idx_threads_forum ON threads (forum);

CREATE INDEX IF NOT EXISTS idx_posts_forum ON posts (forum);
CREATE INDEX IF NOT EXISTS idx_posts_thread_path ON posts (thread, path);
CREATE INDEX IF NOT EXISTS idx_posts_thread_id ON posts (thread, id);
CREATE INDEX IF NOT EXISTS idx_posts_thread_id0 ON posts (thread, id) WHERE parent = 0;
CREATE INDEX IF NOT EXISTS idx_posts_thread_id_created ON posts (id, created, thread);
CREATE INDEX IF NOT EXISTS idx_posts_thread_path1_id ON posts (thread, (path[1]), id);

CREATE INDEX IF NOT EXISTS idx_slug_thread on threads (slug);
CREATE UNIQUE INDEX IF NOT EXISTS idx_votes_thread_nickname ON votes (thread, nick);

-- CREATE INDEX IF NOT EXISTS idx_cover_thread on threads(nick, created, forum, id, message, slug, title, votes);


-- truncate table votes RESTART IDENTITY cascade;truncate table posts RESTART IDENTITY cascade ;truncate table forums RESTART IDENTITY cascade ;truncate table threads RESTART IDENTITY cascade ;truncate table users RESTART IDENTITY cascade ;truncate table forum_users RESTART IDENTITY cascade ;
-- --
-- DROP TABLE if exists users cascade;
-- DROP TABLE if exists forums cascade;
-- DROP TABLE if exists threads cascade;
-- DROP TABLE if exists posts cascade;
-- DROP TABLE if exists votes;
-- DROP TABLE if exists forum_users cascade;
