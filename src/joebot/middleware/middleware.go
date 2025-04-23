package middleware

import (
	"context"
	"fmt"

	"github.com/harmonicinc-com/joebot/repository"
	"github.com/labstack/echo"
)

func IPWhitelist(userRepo repository.UserRepository) echo.MiddlewareFunc {
	whitelist := []string{}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Path() == "/api/login" {
				return next(c)
			}

			token, err := c.Cookie("authToken")
			if err != nil {
				return echo.ErrUnauthorized
			}

			whitelist = fetchWhitelistFromDB(token.Value, userRepo)
			if len(whitelist) == 0 {
				return echo.ErrForbidden
			}

			return echo.ErrForbidden
		}
	}
}

func fetchWhitelistFromDB(token string, userRepo repository.UserRepository) []string {
	IPWhitelist, err := userRepo.GetUserIPWhitelist(context.Background(), token)
	if err != nil {
		fmt.Println("Error fetching IP whitelist:", err)
		return []string{}
	}

	return IPWhitelist
}
