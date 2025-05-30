#!/bin/bash

# Создание базы данных и пользователя
docker exec -i postgres-sentry psql -U sentry -d sentry <<EOF
CREATE DATABASE "sentry";
CREATE ROLE "sentry" WITH LOGIN PASSWORD 'sentry';
GRANT ALL PRIVILEGES ON DATABASE "sentry" TO "sentry";
ALTER USER "sentry" WITH SUPERUSER;
EOF

echo "База данных и пользователь успешно созданы."

# Выполнение миграций для Sentry
docker exec -i postgres-sentry sentry upgrade

echo "Миграции базы данных выполнены успешно."
