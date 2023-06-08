FROM golang:1.20 AS build

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -race -o /app/ddrv ./cmd/ddrv

FROM gcr.io/distroless/base-debian11

WORKDIR /app

COPY --from=build /app/ddrv /app/ddrv

ENTRYPOINT ["/app/ddrv"]