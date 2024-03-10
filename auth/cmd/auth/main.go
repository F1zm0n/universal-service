package main

import "github.com/F1zm0n/uni-auth/repository/postgres"

func main() {
	postgres := postgres.MustNewPostgresDB()
	postgres.MustMigrateSchema()
}
