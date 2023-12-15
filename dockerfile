FROM golang:latest
WORKDIR coffee-shop/
COPY ./ ./

ENV PORT=8080

RUN go build -o main .
CMD ["./main"]