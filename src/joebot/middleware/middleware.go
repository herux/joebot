package middleware

import (
	"net/http"

	"github.com/labstack/echo"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "missing or invalid token",
			})
		}

		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// if token != "valid-token" {
		// 	return c.JSON(http.StatusUnauthorized, map[string]string{
		// 		"error": "unauthorized",
		// 	})
		// }

		// token value in this context is always valid
		c.Set("token", token)

		return next(c)
	}
}
