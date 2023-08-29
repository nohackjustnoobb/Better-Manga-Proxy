FROM golang:1-alpine as builder

WORKDIR /app

COPY . .
RUN go mod download

RUN go build -C cmd -o ../main

FROM alpine:3 as final
WORKDIR /app

COPY --from=builder /app/main /app/main
RUN chmod +x main

EXPOSE 8080
CMD [ "./main" ]
