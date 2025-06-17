FROM golang:1.24.1-bookworm AS build
WORKDIR /app

COPY . .

# Download go modules
RUN go mod download && go mod verify

RUN go build -v -o /go-api ./

# FROM gcr.io/distroless/static-debian12
FROM debian:bookworm-slim

COPY --from=build /go-api /

EXPOSE 8080

CMD ["/go-api"]