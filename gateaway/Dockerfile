FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . ./
RUN go mod download 

RUN GOOS=linux GOARCH=amd64 go build -o /app/gate /app/cmd/gateaway/

RUN chmod +x /app/gate

FROM alpine:latest AS runner

RUN mkdir /app/

COPY --from=builder /app/gate /app
# COPY --from=builder /app/config /app/config

CMD [ "/app/gate" ]

