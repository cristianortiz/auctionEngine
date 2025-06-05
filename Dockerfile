#multistage dockerfile
FROM golang:1.24-alpine AS builder
#upx comprise image size
RUN apk add --no-cache git upx
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
#copy source code
COPY . ./

#build go app
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o auctionengine ./cmd/main.go 
#compress binary
RUN upx --best --lzma auctionengine

FROM alpine:3.21
RUN apk update --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/auctionengine .
COPY .env .env
# midgrations directory
COPY internal/shared/db/migrations/sql ./internal/shared/db/migrations/sql

ENTRYPOINT [ "./auctionengine" ]