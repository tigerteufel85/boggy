FROM golang:latest as builder
ENV GO111MODULE=on
WORKDIR /boggy/
COPY . ./
RUN make dep \
 && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=readonly -a -o /app .

FROM debian:stable-slim
RUN apt-get update -y && \
    apt-get install ca-certificates netcat strace wget -y
RUN update-ca-certificates
RUN mkdir -p images
RUN mkdir -p config
COPY --from=builder app .
CMD ["./app"]