package middleware

import (
	"context"

	"github.com/harmonicinc-com/joebot/repository"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
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
			clientIP := c.RealIP()
			if clientIP == "::1" {
				clientIP = "127.0.0.1"
			}
			for _, allowedIP := range whitelist {
				if clientIP == allowedIP {
					return next(c)
				}
			}

			return echo.ErrForbidden
		}
	}
}

func fetchWhitelistFromDB(token string, userRepo repository.UserRepository) []string {
	user, err := userRepo.GetUserByToken(context.Background(), token)
	if err != nil {
		logrus.Errorf("Error fetching user by token: %v", err)
		return []string{}
	}
	if user == nil {
		logrus.Warnf("No user found for token: %s", token)
		return []string{}
	}

	IPWhitelist, err := userRepo.GetUserIPWhitelist(context.Background(), user.Username)
	if err != nil {
		logrus.Errorf("Error fetching IP whitelist: %v", err)
		return []string{}
	}

	return IPWhitelist
}
