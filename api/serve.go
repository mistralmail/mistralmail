package api

import (
	jwt "github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mistralmail/mistralmail/backend"
	log "github.com/sirupsen/logrus"
)

// jwtCustomClaims represents the claims used in JWT tokens.
type jwtCustomClaims struct {
	UserID uint `json:"userID"`
	jwt.RegisteredClaims
}

// API contains the MistralMail REST API.
type API struct {
	backend *backend.Backend
	config  Config
}

// New creates a new API.
func New(config Config, backend *backend.Backend) (*API, error) {
	return &API{
		backend: backend,
		config:  config,
	}, nil
}

// Serve the API.
func (api *API) Serve() error {
	// Create an Echo instance.
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.CORS())

	e.POST("/auth/login", api.loginHandler)

	g := e.Group("/api")

	// Middleware to secure routes with JWT.
	g.Use(echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwtCustomClaims)
		},
		SigningKey: api.config.Secret,
	}))

	// Users
	g.GET("/users", api.getAllUsersHandler)
	g.DELETE("/users/:id", api.deleteUserHandler)
	g.POST("/users", api.createNewUserHandler)
	g.POST("/reset-password", api.resetPasswordHandler)

	// Metrics
	g.GET("/metrics", api.metricsJSONHandler)

	// Web UI
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  "./web-ui/dist",
		Index: "index.html",
		HTML5: true,
		//Filesystem: http.FS(staticFS),
	}))

	// Start the Echo server.
	log.Printf("Starting API at %s", api.config.HTTPAddress)

	return e.Start(api.config.HTTPAddress)
}
