FROM golang:alpine as builder
ARG VERSION

WORKDIR /src
RUN go env -w GOCACHE=/cache/go-cache
RUN go env -w GOMODCACHE=/cache/gomod-cache
COPY ./go.* ./
RUN --mount=type=cache,target=/gomod-cache \
  go mod download
COPY . .

RUN  --mount=type=cache,target=/cache/gomod-cache --mount=type=cache,target=/cache/go-cache \
    CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o app -ldflags "-X 'main.version=${VERSION}'" ./cmd/subs-api

CMD ["./app", "show-config"]