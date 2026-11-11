FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/servicemeow .

FROM alpine:3.24 AS release
RUN apk add --no-cache ca-certificates
COPY --from=build /out/servicemeow /servicemeow

USER 65534:65534
ENTRYPOINT [ "/servicemeow" ]
CMD ["--help"]
