FROM golang:1.15.6-alpine

#d# setup directories
RUN mkdir /app
ADD . /app
WORKDIR /app

## pull dependencies
RUN go mod download

## build app
RUN go build -o main .

## start application
CMD ["/app/main"]