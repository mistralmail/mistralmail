FROM --platform=$BUILDPLATFORM golang:latest AS build

ARG TARGETARCH

WORKDIR /random_work_dir

# first download dependencies
COPY go.mod /random_work_dir
COPY go.sum /random_work_dir
RUN go mod download

# then copy source code
COPY / /random_work_dir


RUN GOOS=linux GOARCH=$TARGETARCH CGO_ENABLED=1 go build -o /random_work_dir/mistralmail ./cmd/mistralmail


FROM golang:latest

WORKDIR /

COPY --from=build --chown=${USERNAME}:${USERNAME} /random_work_dir/mistralmail /mistralmail/

WORKDIR /mistralmail

RUN chmod +x ./mistralmail

RUN mkdir ./certificates

EXPOSE 80
EXPOSE 443
EXPOSE 143
EXPOSE 587
EXPOSE 25

CMD ["./mistralmail"]