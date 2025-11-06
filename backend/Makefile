run:
ADDR=:8080 DATABASE_URL=postgres://postgres:postgres@localhost:5432/shudong?sslmode=disable go run ./cmd/api


dev-db:
docker compose up -d db


up:
docker compose up --build


fmt:
gofmt -w .