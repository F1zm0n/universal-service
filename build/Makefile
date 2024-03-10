AUTH_BIN=authApp

auth:
	@cd ../auth && env GOOS=linux CGO_ENABLED=0 go build -o ${AUTH_BIN} ./cmd/auth 
	@../auth/${AUTH_BIN}

bup: 
	@docker-compose down
	@docker-compose up -d --build
