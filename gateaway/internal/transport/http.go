package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type VerifyPayload struct {
	VerId uuid.UUID `json:"ver_id"`
}
type ErrResponse struct {
	Err error `json:"error"`
}

func HandleRegister(c echo.Context) error {
	req, err := http.NewRequest(http.MethodPost, "http://producer:5000/mail", c.Request().Body)
	if err != nil {
		return err
	}
	defer c.Request().Body.Close()
	cli := http.DefaultClient
	res, err := cli.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("status code is not 200")
	}
	var dat []byte
	_, err = res.Body.Read(dat)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return c.JSON(http.StatusOK, map[string]any{"error": nil})
}

func HandleVerify(c echo.Context) error {
	id := c.QueryParam("id")
	verId, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	data := VerifyPayload{
		VerId: verId,
	}
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, "http://producer:5000/verify", bytes.NewReader(j))
	if err != nil {
		return err
	}
	defer c.Request().Body.Close()
	cli := http.DefaultClient
	res, err := cli.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("status code is not 200")
	}
	return c.JSON(http.StatusOK, map[string]any{"error": nil})
}

// func HandleLogin(c echo.Context) error {
// 	req, err := http.NewRequest(http.MethodPost, "http://mailer:5001/mail", c.Request().Body)
// 	if err != nil {
// 		return err
// 	}
// }
