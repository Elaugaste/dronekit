FROM golang:1.22-alpine AS build
RUN mkdir /src
WORKDIR /src
COPY . /src/
RUN apk add gcc make build-base
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -o /app_server .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=arm
RUN go build -o /app_client .

FROM alpine
WORKDIR /
COPY --from=build /app_server /app
COPY --from=build /app_client /client
RUN mkdir /db