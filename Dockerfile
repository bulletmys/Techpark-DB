FROM golang:1.14 AS builder
ENV GO111MODULE=on
WORKDIR /app

COPY go.mod .
RUN go mod download

COPY . .
RUN go build main.go


# Postgres
FROM ubuntu:18.04 AS release
ENV DEBIAN_FRONTEND=noninteractive
ENV PGVER 12

RUN apt-get update && apt-get install -y wget gnupg && \
    wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
RUN echo "deb http://apt.postgresql.org/pub/repos/apt bionic-pgdg main" > /etc/apt/sources.list.d/PostgreSQL.list

RUN apt -y update && apt install -y postgresql-$PGVER

WORKDIR DB
COPY . .

USER postgres
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    psql -d docker -c "CREATE EXTENSION IF NOT EXISTS citext;" &&\
    psql -d docker -c "set synchronous_commit to 'off';" &&\
    psql -d docker -c "set random_page_cost to '1';" &&\
#    psql -d docker -c "ALTER SYSTEM SET max_connections = '200'; ALTER SYSTEM SET shared_buffers = '256MB'; ALTER SYSTEM SET effective_cache_size = '768MB'; ALTER SYSTEM SET maintenance_work_mem = '64MB'; ALTER SYSTEM SET checkpoint_completion_target = '0.7'; ALTER SYSTEM SET wal_buffers = '7864kB'; ALTER SYSTEM SET default_statistics_target = '100'; ALTER SYSTEM SET random_page_cost = '1.1'; ALTER SYSTEM SET effective_io_concurrency = '200'; ALTER SYSTEM SET work_mem = '1310kB'; ALTER SYSTEM SET min_wal_size = '1GB'; ALTER SYSTEM SET max_wal_size = '4GB'; ALTER SYSTEM SET max_worker_processes = '2'; ALTER SYSTEM SET max_parallel_workers_per_gather = '1'; ALTER SYSTEM SET max_parallel_workers = '2'; ALTER SYSTEM SET max_parallel_maintenance_workers = '1';" &&\
#    psql -d docker -c "alter user docker set shared_buffers='256MB';" &&\
    psql docker -a -f sql/init.sql &&\
    /etc/init.d/postgresql stop

RUN echo "local all all md5" > /etc/postgresql/$PGVER/main/pg_hba.conf && \
    echo "host all all 0.0.0.0/0 md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf

RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf

EXPOSE 5432
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]
USER root
EXPOSE 5000

COPY --from=builder /app/main .

CMD service postgresql start && ./main