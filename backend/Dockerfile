FROM golang:1.21-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org,direct
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/unalone .

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/unalone /usr/local/bin/unalone
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["/usr/local/bin/unalone"]
