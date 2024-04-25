FROM golang:1.22-alpine AS builder

RUN apk add alpine-sdk 

WORKDIR /app

COPY . ./
RUN go mod download 

RUN GOOS=linux GOARCH=amd64 go build -tags musl -o /app/consumer /app/cmd/mailer/

RUN chmod +x /app/consumer

FROM alpine:latest AS runner

RUN mkdir /app/

COPY --from=builder /app/consumer /app

CMD [ "/app/consumer" ]
