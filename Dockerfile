FROM golang:1.14-alpine as build

WORKDIR /go/src/servicemeow
COPY . .

RUN go get -d -v ./...
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-s -w" -o servicemeow

FROM alpine:3.12.0 as release
COPY --from=build /go/src/servicemeow/servicemeow ./servicemeow

ENTRYPOINT [ "/servicemeow" ]
CMD ["--help"]
