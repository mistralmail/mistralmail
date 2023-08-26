package api

import (
	"net/http"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (api *API) loginHandler(c echo.Context) error {
	// Parse the request to get the email and password.
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Check against bruteforce
	canLogin, err := api.backend.LoginAttempts.CanLogin(c.RealIP())
	if err != nil {
		log.Errorf("couldn't check login attempts: %v", err)
		return echo.ErrInternalServerError
	}
	if !canLogin {
		log.Errorf("login attempts exceeded for: %s", c.RealIP())
		return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
			"error": "Too many failed login attempts",
		})
	}

	// Authenticate the user by querying your UserRepository.
	user, err := api.backend.UserRepo.FindUserByEmail(email)
	if err != nil {
		_, err := api.backend.LoginAttempts.AddFailedAttempts(c.RealIP())
		if err != nil {
			log.Errorf("couldn't increase login attempts: %v", err)
		}
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Email or password incorrect",
		})
	}
	validLogin, err := user.CheckPassword(password)
	if err != nil {
		return echo.ErrInternalServerError
	}
	if !validLogin {
		_, err := api.backend.LoginAttempts.AddFailedAttempts(c.RealIP())
		if err != nil {
			log.Errorf("couldn't increase login attempts: %v", err)
		}
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Email or password incorrect",
		})
	}

	// Create a JWT token.

	claims := &jwtCustomClaims{
		UserID: user.ID, // Customize the claims as needed.
		RegisteredClaims: jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token and generate a JWT token string.
	tokenString, err := token.SignedString(api.config.Secret) // Change to your actual secret key.
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": tokenString,
		"user":  user,
	})
}
