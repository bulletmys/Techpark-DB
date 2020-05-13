create extension if not exists citext;
CREATE TABLE users
(
    nick  CITEXT COLLATE ucs_basic PRIMARY KEY,
    email CITEXT UNIQUE,
    name  CITEXT NOT NULL,
    about VARCHAR
);

CREATE TABLE forums
(
    slug    CITEXT PRIMARY KEY,
    nick    CITEXT  NOT NULL,
    title   VARCHAR NOT NULL,
    threads INT    DEFAULT 0,
    posts   BIGINT DEFAULT 0,
    FOREIGN KEY (nick) REFERENCES users (nick)
);

CREATE TABLE threads
(
    nick    CITEXT  NOT NULL,
    created timestamptz,
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

CREATE TABLE posts
(
    nick     CITEXT  NOT NULL,
    created  timestamptz,
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

CREATE TABLE votes
(
    nick   CITEXT not null,
    vote   bool   not null,
    thread int4   not null,
    FOREIGN KEY (nick) REFERENCES users (nick),
    FOREIGN KEY (thread) REFERENCES threads (id)
);

select nick,
       created,
       forum,
       id,
       message,
       slug,
       title,
       votes
from threads
where forum = 'V3KHZG5ww33as'
order by created
limit 100;

select lower(nick), name, email, about
from users
where lower(nick) = lower('HAC.fzCNeqpWi9k9jD');
select nick, name, email, about
from users
where lower(nick) = lower('hac.FZcnEQPwI9K9Jd');

select slug, nick, title, threads, posts
from forums
where lower(slug) = 'lgoaikigymm3r';

select created, id, path
from posts
where thread = 937
  and id > 793
order by path;

CREATE OR REPLACE FUNCTION check_global_var(v bigint, cur bigint) RETURNS bool as
$$
DECLARE
    p int4;
BEGIN
    p := (SELECT current_setting('myapp.var1'));
    if (cur > v and p = '1') then
        SELECT set_config('myapp.var1', '0', false);
        return (select true);
    else
        return (select false);
    end if;
END;
$$ LANGUAGE plpgsql;

-- CREATE OR REPLACE FUNCTION check_global_var(v bigint, cur bigint) RETURNS bool as
-- $$
-- BEGIN;
--     set p = (SELECT current_setting('myapp.var1'));
--     IF cur > v and p = '1' then
--         SELECT set_config('myapp.var1', '0', True);
--         return true;
--     else
--         return false;
--     end if;
-- END;
-- $$ LANGUAGE SQL;

start transaction;

SELECT set_config('myapp.var1', '1', false);

select path,
       id,
       ROW_NUMBER() OVER (ORDER BY id)
from posts
where thread = 2956
  and id > 3598
order by path, id;

COMMIT;

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

SELECT *
FROM posts
WHERE thread = 83
order by path, id;

SELECT *
FROM posts
WHERE thread = 83
  and path[1] = id
order by path, id;

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


SELECT *
FROM get_all_foo(4241, 1634);

SELECT *
FROM get_parent_tree(65, 89, 0);

SELECT *
FROM posts
WHERE parent = 0
order by id desc;

CREATE OR REPLACE FUNCTION lol() RETURNS posts AS
$BODY$
DECLARE
    r   posts%rowtype;
    r2  RECORD;
    kek int;
BEGIN
    kek := 0;
    FOR r IN
        SELECT * FROM posts WHERE parent = 0 order by id desc
        LOOP
            RETURN (select * from posts where parent = r.parent order by id);
        END LOOP;
END
$BODY$
    LANGUAGE plpgsql;


with sincePost AS (
    select *
    from posts
    where id = 19605
)
SELECT p2.nick, p2.created, p2.forum, p2.id, p2.message, p2.thread, p2.path
FROM (select *
      from posts
      WHERE parent = 0
        and thread = 1503
        and id < (select sincePost.path[1] from sincePost)
      order by id desc
      limit 3) p1
         join posts p2 on (p1.id = p2.path[1] or p1.id = p2.id)
order by p1.id desc, p2.path;

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


SELECT *
FROM get_all_foo2(3920, 392);

select *
from posts
where thread = 392
order by path desc, id
offset 3 limit 3;


truncate table votes RESTART IDENTITY cascade;truncate table posts RESTART IDENTITY cascade ;truncate table forums RESTART IDENTITY cascade ;truncate table threads RESTART IDENTITY cascade ;truncate table users RESTART IDENTITY cascade ;

DROP TABLE if exists users cascade;
DROP TABLE if exists forums cascade;
DROP TABLE if exists threads cascade;
DROP TABLE if exists posts cascade;
DROP TABLE if exists votes;

select *
from (
         select u.nick, u.name, u.email, u.about
         from forums f
                  join threads t on lower(f.slug) = lower(t.forum)
                  join posts p on lower(f.slug) = lower(p.forum)
                  join users u on (lower(t.nick) = lower(u.nick) or lower(p.nick) = lower(u.nick))
         where lower(f.slug) = lower('mmYK5NT4hLM3s')
         group by u.nick
     ) as ftpu
order by lower(nick);

select *
from (
         select u.nick, u.name, u.email, u.about
         from forums f
                  full join posts p on lower(f.slug) = lower(p.forum)
                  full join threads t on lower(f.slug) = lower(t.forum)
                  join users u on (lower(t.nick) = lower(u.nick) or lower(p.nick) = lower(u.nick))
         where lower(f.slug) = lower('C1B5zF04_3lL8V')
         group by u.nick) as kek
order by lower(nick);

--and lower(u.nick) < lower('T5gRHOOGB3LkP.Jill')

CREATE FUNCTION btrsort_nextunit(text) RETURNS text AS
$$
SELECT CASE
           WHEN $1 ~ '^[^0-9]+' THEN
               COALESCE(SUBSTR($1, LENGTH(SUBSTRING($1 FROM '[^0-9]+')) + 1), '')
           ELSE
               COALESCE(SUBSTR($1, LENGTH(SUBSTRING($1 FROM '[0-9]+')) + 1), '')
           END

$$ LANGUAGE SQL;

CREATE FUNCTION btrsort(text) RETURNS text AS
$$
SELECT CASE
           WHEN char_length($1) > 0 THEN
               CASE
                   WHEN $1 ~ '^[^0-9]+' THEN
                           RPAD(SUBSTR(COALESCE(SUBSTRING($1 FROM '^[^0-9]+'), ''), 1, 12), 12, ' ') ||
                           btrsort(btrsort_nextunit($1))
                   ELSE
                           LPAD(SUBSTR(COALESCE(SUBSTRING($1 FROM '^[0-9]+'), ''), 1, 12), 12, ' ') ||
                           btrsort(btrsort_nextunit($1))
                   END
           ELSE
               $1
           END
    ;
$$ LANGUAGE SQL;
