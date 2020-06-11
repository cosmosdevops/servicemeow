FROM golang:1.14

WORKDIR /go/src/servicemeow
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...
ENTRYPOINT [ "servicemeow" ]
CMD ["--help"]
