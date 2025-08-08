# syntax=docker/dockerfile:1

FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/producer ./cmd/producer

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/producer /app/producer
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/app/producer"] 