FROM golang:1.13

ENV GOOS=linux
ENV GOARCH=amd64

EXPOSE 8080

WORKDIR /go-kubectx

COPY go.mod go.sum /go-kubectx/
RUN go mod download

COPY . .