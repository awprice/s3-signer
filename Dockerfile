FROM golang:1.10 as builder
WORKDIR /go/src/github.com/awprice/s3-signer/
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
ADD ./ ./
RUN make setup
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o s3-signer main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/src/github.com/awprice/s3-signer/s3-signer .
CMD [ "./s3-signer" ]