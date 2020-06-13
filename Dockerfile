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
#    psql -d docker -c "set synchronous_commit to 'off';" &&\
    psql -d docker -c "set random_page_cost to '1';" &&\
#    psql -d docker -c "alter user docker set shared_buffers='256MB';" &&\
    psql docker -a -f sql/init.sql &&\
    /etc/init.d/postgresql stop

RUN echo "local all all md5" > /etc/postgresql/$PGVER/main/pg_hba.conf && \
    echo "host all all 0.0.0.0/0 md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf

#RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf
#
#RUN echo "synchronous_commit='off'" >> /etc/postgresql/$PGVER/main/postgresql.conf
#RUN echo "fsync = 'off'" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "synchronous_commit='off'" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "fsync = 'off'" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "work_mem = 16MB" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "effective_cache_size = 1024MB" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "shared_buffers = 256MB" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "max_wal_size = 1GB" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "logging_collector = 'off'" >> /etc/postgresql/$PGVER/main/postgresql.conf

EXPOSE 5432
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]
USER root
EXPOSE 5000

COPY --from=builder /app/main .

CMD service postgresql start && ./main