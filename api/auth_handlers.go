package api

import (
	"net/http"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func (api *API) loginHandler(c echo.Context) error {
	// Parse the request to get the email and password.
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Authenticate the user by querying your UserRepository.
	user, err := api.backend.UserRepo.FindUserByEmail(email)
	if err != nil {
		return echo.ErrUnauthorized
	}
	canLogin, err := user.CheckPassword(password)
	if err != nil {
		return echo.ErrInternalServerError
	}
	if !canLogin {
		return echo.ErrUnauthorized
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
