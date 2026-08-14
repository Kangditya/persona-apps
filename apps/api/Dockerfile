FROM golang:1.26.6-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /persona-api ./cmd/api

FROM alpine:3.23
RUN adduser -D -u 10001 persona
USER persona
COPY --from=build /persona-api /persona-api
EXPOSE 8080
ENTRYPOINT ["/persona-api"]
