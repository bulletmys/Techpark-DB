create extension if not exists citext;
CREATE UNLOGGED TABLE users
(
    nick  CITEXT COLLATE ucs_basic PRIMARY KEY,
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
    nick    CITEXT  NOT NULL REFERENCES users (nick),
    created timestamptz,
    forum   CITEXT  NOT NULL REFERENCES forums (slug),
    id      BIGSERIAL PRIMARY KEY,
    message VARCHAR NOT NULL,
    slug    CITEXT UNIQUE,
    title   VARCHAR NOT NULL,
    votes   int default 0
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
    nick     CITEXT  NOT NULL REFERENCES users (nick),
    created  timestamptz,
    forum    CITEXT  NOT NULL REFERENCES forums (slug),
    id       BIGSERIAL PRIMARY KEY,
    isEdited bool    NOT NULL,
    message  VARCHAR NOT NULL,
    parent   BIGINT DEFAULT 0,
    thread   int     not null REFERENCES threads (id),
    path     BIGINT[]
);

CREATE OR REPLACE FUNCTION after_post_insert_func() RETURNS TRIGGER AS
$after_post_insert$
DECLARE
    parentArray BIGINT[];
BEGIN
    update forums set posts = posts + 1 where NEW.forum = slug;
    IF (NEW.parent <> 0) THEN
--         parentArray := (select path from posts where id = NEW.parent);
        NEW.path = NEW.path || currval('posts_id_seq');
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
    nick   CITEXT not null REFERENCES users (nick),
    vote   bool   not null,
    thread int4   not null REFERENCES threads (id)
);

CREATE OR REPLACE FUNCTION get_all_foo(since int, threadID bigint) RETURNS int AS
$BODY$
DECLARE
    r   posts%rowtype;
    kek int;
BEGIN
    kek := 0;
    FOR r IN
        SELECT * FROM posts WHERE thread = threadID order by path, id
        LOOP
            kek = kek + 1;
            if r.id = since then
                RETURN kek;
            end if;
        END LOOP;
    RAISE NOTICE '%', kek;
    RETURN kek;
END
$BODY$
    LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION get_parent_tree(limiting int, threadID bigint, dataOffset int) RETURNS SETOF posts AS
$BODY$
DECLARE
    r   posts%rowtype;
    kek int;
BEGIN
    kek := 0;
    FOR r IN
        SELECT * FROM posts WHERE thread = threadID order by path, id offset dataOffset
        LOOP
            if r.parent = 0 then
                kek = kek + 1;
            end if;
            if kek > limiting then
                return;
            end if;
            RETURN next r;
        END LOOP;
    RETURN;
END
$BODY$
    LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_all_foo2(since int, threadID bigint) RETURNS int AS
$BODY$
DECLARE
    r   posts%rowtype;
    kek int;
BEGIN
    kek := 0;
    FOR r IN
        SELECT * FROM posts WHERE thread = threadID order by path desc, id
        LOOP
            kek = kek + 1;
            if r.id = since then
                RETURN kek;
            end if;
        END LOOP;
    RAISE NOTICE '%', kek;
    RETURN kek;
END
$BODY$
    LANGUAGE plpgsql;


CREATE UNIQUE INDEX IF NOT EXISTS index_forum_slug  ON forums (slug);
CREATE INDEX IF NOT EXISTS index_posts_forum ON posts (thread);
-- CREATE INDEX IF NOT EXISTS index_posts_forum_parent ON posts (parent); -- ???
CREATE INDEX IF NOT EXISTS index_threads_forum ON threads (forum);
CREATE INDEX IF NOT EXISTS index_posts_thread_id_parent ON posts (thread, id) WHERE parent = 0;
CREATE INDEX IF NOT EXISTS index_posts_thread_id ON posts (id, thread);
-- CREATE INDEX IF NOT EXISTS index_posts_thread_path_first ON posts (thread, (path[1]), id);
CREATE INDEX IF NOT EXISTS index_threads_slug ON threads (slug);
-- CREATE INDEX IF NOT EXISTS index_posts_thread_id_triple ON posts (id, created, thread);
CREATE UNIQUE INDEX IF NOT EXISTS index_votes_thread_nickname ON votes (thread, nick);
CREATE INDEX IF NOT EXISTS index_posts_thread_path ON posts (thread, path);

--
-- truncate table votes RESTART IDENTITY cascade;truncate table posts RESTART IDENTITY cascade ;truncate table forums RESTART IDENTITY cascade ;truncate table threads RESTART IDENTITY cascade ;truncate table users RESTART IDENTITY cascade ;
--
-- DROP TABLE if exists users cascade;
-- DROP TABLE if exists forums cascade;
-- DROP TABLE if exists threads cascade;
-- DROP TABLE if exists posts cascade;
-- DROP TABLE if exists votes;
--
-- SELECT
--   relname,
--   seq_scan - idx_scan AS too_much_seq,
--   CASE
--     WHEN
--       seq_scan - coalesce(idx_scan, 0) > 0
--     THEN
--       'Missing Index?'
--     ELSE
--       'OK'
--   END,
--   pg_relation_size(relname::regclass) AS rel_size, seq_scan, idx_scan
-- FROM
--   pg_stat_all_tables
-- WHERE
--   schemaname = 'public'
--   AND pg_relation_size(relname::regclass) > 80000
-- ORDER BY
--   too_much_seq DESC;
--
-- SELECT
--   indexrelid::regclass as index,
--   relid::regclass as table,
--   'DROP INDEX ' || indexrelid::regclass || ';' as drop_statement
-- FROM
--   pg_stat_user_indexes
--   JOIN
--     pg_index USING (indexrelid)
-- WHERE
--   idx_scan = 0
--   AND indisunique is false;

ALTER SYSTEM SET
    max_connections = '200';
ALTER SYSTEM SET
    shared_buffers = '256MB';
ALTER SYSTEM SET
    effective_cache_size = '768MB';
ALTER SYSTEM SET
    maintenance_work_mem = '64MB';
ALTER SYSTEM SET
    checkpoint_completion_target = '0.7';
ALTER SYSTEM SET
    wal_buffers = '7864kB';
ALTER SYSTEM SET
    default_statistics_target = '100';
ALTER SYSTEM SET
    random_page_cost = '1.1';
ALTER SYSTEM SET
    effective_io_concurrency = '200';
ALTER SYSTEM SET
    work_mem = '1310kB';
ALTER SYSTEM SET
    min_wal_size = '1GB';
ALTER SYSTEM SET
    max_wal_size = '4GB';
ALTER SYSTEM SET
    max_worker_processes = '2';
ALTER SYSTEM SET
    max_parallel_workers_per_gather = '1';
ALTER SYSTEM SET
    max_parallel_workers = '2';
ALTER SYSTEM SET
    max_parallel_maintenance_workers = '1';
