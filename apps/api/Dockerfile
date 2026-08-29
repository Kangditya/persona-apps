FROM golang:1.26.6-alpine AS build

WORKDIR /src/apps/api
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api .
COPY contracts/openapi /src/contracts/openapi
RUN CGO_ENABLED=0 go build -o /persona-api ./cmd/api

FROM alpine:3.23
RUN adduser -D -u 10001 persona
WORKDIR /app
USER persona
COPY --from=build /persona-api /persona-api
COPY --from=build /src/contracts/openapi /app/contracts/openapi
EXPOSE 8080
ENTRYPOINT ["/persona-api"]
