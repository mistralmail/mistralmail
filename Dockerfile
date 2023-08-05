FROM golang:alpine

WORKDIR /app

COPY gopistolet /app/

ENTRYPOINT ["/app/gopistolet"]
