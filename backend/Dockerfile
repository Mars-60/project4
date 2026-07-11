FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/tradepilot ./cmd/server

FROM alpine:3.20
RUN adduser -D -H tradepilot
WORKDIR /app
COPY --from=build /out/tradepilot /app/tradepilot
COPY migrations /app/migrations
USER tradepilot
EXPOSE 8080
ENTRYPOINT ["/app/tradepilot"]
