FROM --platform=$BUILDPLATFORM golang:latest AS build

ARG TARGETARCH

WORKDIR /random_work_dir

# first download dependencies
COPY go.mod /random_work_dir
COPY go.sum /random_work_dir
RUN go mod download

# then copy source code
COPY / /random_work_dir

# Build server
RUN GOOS=linux GOARCH=$TARGETARCH CGO_ENABLED=1 go build -o /random_work_dir/mistralmail ./cmd/mistralmail

# Build CLI
RUN GOOS=linux GOARCH=$TARGETARCH CGO_ENABLED=1 go build -o /random_work_dir/mistralmail-cli ./cmd/mistralmail-cli


FROM node:16 as buildWeb

ARG TARGETARCH

WORKDIR /random_work_dir/web-ui/

RUN npm install -g vite

COPY /web-ui /random_work_dir/web-ui


RUN yarn install

RUN yarn build

RUN ls /
RUN ls /random_work_dir/
RUN ls /random_work_dir/web-ui



FROM golang:latest

WORKDIR /

COPY --from=build --chown=${USERNAME}:${USERNAME} /random_work_dir/mistralmail /mistralmail/
COPY --from=build --chown=${USERNAME}:${USERNAME} /random_work_dir/mistralmail-cli /mistralmail/

COPY --from=buildWeb --chown=${USERNAME}:${USERNAME} /random_work_dir/web-ui/dist /mistralmail/web-ui/dist

WORKDIR /mistralmail

RUN chmod +x ./mistralmail
RUN chmod +x ./mistralmail-cli

ENV PATH="/mistralmail:${PATH}" 

RUN mkdir ./certificates

EXPOSE 80
EXPOSE 443
EXPOSE 143
EXPOSE 587
EXPOSE 25
EXPOSE 8080

CMD ["./mistralmail"]