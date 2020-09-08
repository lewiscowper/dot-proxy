FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git
WORKDIR /tmp/build

COPY go.mod go.sum ./
RUN go mod download \
    && go mod verify

COPY ./ ./

RUN CGO_ENABLED=0 \
  go build -o dotproxy -ldflags='-w -s -extldflags "-static"' /tmp/build/

FROM gcr.io/distroless/static

COPY --from=builder /tmp/build/dotproxy /

EXPOSE ${DOTPROXY_LISTEN_PORT}

CMD ["/dotproxy"]
