FROM golang:1.8

WORKDIR /go/src/app
COPY ./main.go .
COPY ./main_test.go .
COPY ./test_data.csv .

RUN go get -d -v ./...
RUN go install -v ./...
RUN CGO_ENABLED=1

CMD ["app"]