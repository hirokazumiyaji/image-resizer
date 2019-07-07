FROM golang:1.12 as builder

WORKDIR /go/src/github.com/hirokazumiyaji/image-resizer
COPY . .

ENV GO111MODULE "on"

RUN CGO_ENABLED=0 GOOS=linux go build -v -o image-resizer

FROM alpine
RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/github.com/hirokazumiyaji/image-resizer /image-resizer

ENV PORT 8080

CMD ["/image-resizer"]
