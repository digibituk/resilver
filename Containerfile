FROM docker.io/library/golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /resilver ./cmd/resilver

FROM scratch
COPY --from=builder /resilver /resilver
COPY configs/default.json /etc/resilver/config.json
EXPOSE 8080
ENTRYPOINT ["/resilver", "--config", "/etc/resilver/config.json"]
