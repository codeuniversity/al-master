FROM golang:1.11 as builder
WORKDIR /go/src/github.com/MonteyMontey/al-master
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -o master main/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /go/src/github.com/MonteyMontey/al-master/master .
EXPOSE 4000
CMD ["./master"]
