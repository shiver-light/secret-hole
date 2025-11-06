FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod go build -o /out/api ./cmd/api


FROM alpine:3.20
ENV ADDR=":8080"
COPY --from=build /out/api /usr/local/bin/api
EXPOSE 8080
CMD ["/usr/local/bin/api"]