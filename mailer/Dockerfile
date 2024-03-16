FROM golang:1.22-alpine AS builder

RUN apk add alpine-sdk 

WORKDIR /app

COPY . ./
RUN go mod download 

RUN GOOS=linux GOARCH=amd64 go build -o /app/mailer /app/cmd/mailer/

RUN chmod +x /app/mailer

FROM alpine:latest AS runner

RUN mkdir /app/

COPY --from=builder /app/mailer /app
COPY --from=builder /app/config /app/config

CMD [ "/app/mailer" ]
