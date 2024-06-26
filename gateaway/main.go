package main

import (
	"log"

	"github.com/labstack/echo/v4"

	"github.com/F1zm0n/universal-gateaway/internal/transport"
)

func main() {
	var (
		e      = echo.New()
		auth   = e.Group("/a")
		unauth = e.Group("/u")
	)

	auth.Use(transport.JWTAuthentication)

	unauth.POST("/register", transport.HandleRegister)
	unauth.GET("/verify", transport.HandleVerify)
	unauth.GET("/login", transport.HandleLogin)
	log.Fatal(e.Start(":3000"))
}
