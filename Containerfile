FROM docker.io/library/golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /resilver ./cmd/resilver

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /resilver /resilver
EXPOSE 8080
ENTRYPOINT ["/resilver"]
CMD ["--config", "/config.json"]
