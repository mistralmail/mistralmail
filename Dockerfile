FROM golang:alpine

WORKDIR /app

COPY cmd/mistralmail/mistralmail /app/

ENTRYPOINT ["/app/mistralmail"]
