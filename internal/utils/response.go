package utils

import "github.com/labstack/echo/v4"

func ErrorResponse(c echo.Context, code int, err error) error {
	return c.JSON(code, map[string]string{
		"message": err.Error(),
	})
}
