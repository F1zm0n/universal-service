package transport

import (
	"fmt"
	"log"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func JWTAuthentication(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("X-Api-Token")
		claims, err := ParseToken(token)
		if err != nil {
			log.Println(err)
			return err
		}
		email := claims["email"].(string)
		c.Set("email", email)
		return next(c)
	}
}

func ParseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unauthorized")
		}
		secret := "nigga"
		if secret == "" {
			log.Fatal("no secret is set, set the env secret")
		}
		return []byte(secret), nil
	})
	if err != nil {
		log.Println("failed to parse jwt token: ", err)
		return nil, fmt.Errorf("unauthorized")
	}

	if cl, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return cl, nil
	}

	return nil, fmt.Errorf("unauthorized")
}
