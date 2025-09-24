Пример переменных окружения для инициализации конфига: \

GRPC_PORT=9090;GRPC_GATEWAY_PORT=8080;POSTGRES_HOST=localhost;POSTGRES_PORT=5432;POSTGRES_DB=library;POSTGRES_USER=user;POSTGRES_PASSWORD=1234567;POSTGRES_MAX_CONN=10;OUTBOX_ENABLED=true;OUTBOX_WORKERS=5;OUTBOX_BATCH_SIZE=100;OUTBOX_WAIT_TIME_MS=5000;OUTBOX_IN_PROGRESS_TTL_MS=10000;OUTBOX_BOOK_SEND_URL=http://localhost:8081/books;OUTBOX_AUTHOR_SEND_URL=http://localhost:8081/authors

OUTBOX_BATCH_SIZE определяет количество задач, которые может взять 1 worker. \
OUTBOX_WAIT_TIME определяет время сна между обращениями воркера к бд. \
OUTBOX_IN_PROGRESS_TTL определяет время, через которое задачу возьмет другой воркер. \

