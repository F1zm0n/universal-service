FROM golang:1.22-alpine AS builder

RUN apk add alpine-sdk 

WORKDIR /app

COPY . ./
RUN go mod download 

RUN GOOS=linux GOARCH=amd64 go build -tags musl -o /app/producer /app/cmd/producer/

RUN chmod +x /app/producer

FROM alpine:latest AS runner

RUN mkdir /app/

COPY --from=builder /app/producer /app
COPY --from=builder /app/config /app/config

CMD [ "/app/producer" ]


