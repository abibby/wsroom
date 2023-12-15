FROM golang:1.21 as go-build

WORKDIR /go/src/github.com/abibby/comicbox-3

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /wsroom

# Now copy it into our base image.
FROM alpine

RUN apk update && \
    apk add ca-certificates && \
    update-ca-certificates

COPY --from=go-build /wsroom /wsroom

EXPOSE 3335/tcp

CMD ["/wsroom"]
