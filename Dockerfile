FROM golang:alpine

WORKDIR /app

COPY cmd/gopistolet/gopistolet /app/

ENTRYPOINT ["/app/gopistolet"]
