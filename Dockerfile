FROM golang:1.17-alpine as builder
ENV GO111MODULE=on
WORKDIR /boggy/
COPY . ./

RUN apk add git
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o /app boggy.go

FROM alpine:latest as alpine
RUN apk add --no-cache git ca-certificates tzdata
RUN mkdir -p config
COPY --from=builder app .

CMD ["./app"]