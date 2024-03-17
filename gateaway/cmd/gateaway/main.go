package main

import (
	"github.com/labstack/echo/v4"

	"github.com/F1zm0n/universal-gateaway/internal/transport"
)

func main() {
	e := echo.New()
	e.POST("/register", transport.HandleRegister)
	e.GET("/verify", transport.HandleVerify)
	e.Start(":5002")
}
