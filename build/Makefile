AUTH_BIN=authApp
MAIL_BIN=mailerApp
PRODUCER_BIN=prodApp

auth:
	@cd ../auth && env GOOS=linux CGO_ENABLED=0 go build -o ${AUTH_BIN} ./cmd/auth 
	@../auth/${AUTH_BIN}

prod:
	@cd ../producer && env GOOS=linux CGO_ENABLED=1 go build -o ${PRODUCER_BIN} ./cmd/producer
	@../producer/${PRODUCER_BIN}

bup: 
	@docker-compose down
	@docker-compose up -d --build

