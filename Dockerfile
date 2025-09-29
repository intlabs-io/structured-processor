# Step 1

FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /app/cmd/server

RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Step 2

# Changing to apline fixed it
# Multi stage build results in smaller image

FROM alpine

WORKDIR /app

COPY --from=builder /app/cmd/server/server .

EXPOSE 4040

ENV GIN_MODE=release

ENTRYPOINT ["/app/server"]