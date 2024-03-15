FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . ./
RUN go mod download 

RUN GOOS=linux GOARCH=amd64 go build -o /app/auth /app/cmd/auth/

RUN chmod +x /app/auth

FROM alpine:latest AS runner

RUN mkdir /app/

COPY --from=builder /app/auth /app
COPY --from=builder /app/config /app/config

CMD [ "/app/auth" ]

