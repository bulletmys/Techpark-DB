FROM golang:1.14 AS builder
WORKDIR /app

COPY go.mod .
RUN go mod download

COPY . .
RUN go build main.go


# Postgres
FROM ubuntu:18.04 AS release
ENV DEBIAN_FRONTEND=noninteractive
ENV PGVER 11

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
    psql docker -a -f sql/init.sql &&\
    /etc/init.d/postgresql stop

RUN echo "local all all md5" > /etc/postgresql/$PGVER/main/pg_hba.conf && \
    echo "host all all 0.0.0.0/0 md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf

RUN cat sql/postgresql.conf >> /etc/postgresql/$PGVER/main/postgresql.conf
EXPOSE 5432
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]
USER root
EXPOSE 5000

COPY --from=builder /app/main .

CMD service postgresql start && ./main