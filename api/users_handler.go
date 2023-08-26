package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (api *API) getAllUsersHandler(c echo.Context) error {
	// Call the GetAllUsers method from your UserRepository.
	users, err := api.backend.UserRepo.GetAllUsers()
	if err != nil {
		// Handle the error, e.g., return a JSON response with an error message.
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Return a JSON response with the list of users.
	return c.JSON(http.StatusOK, users)
}

func (api *API) deleteUserHandler(c echo.Context) error {
	// Parse the user ID from the URL parameter.
	userIDParam := c.Param("id")
	userID, err := strconv.ParseUint(userIDParam, 10, 64)
	if err != nil {
		// Handle invalid ID format with a bad request response.
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid user ID format",
		})
	}

	// Call the DeleteUser method from your UserRepository.
	err = api.backend.UserRepo.DeleteUser(uint(userID))
	if err != nil {
		// Handle the error, e.g., return a JSON response with an error message.
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Return a success response.
	return c.JSON(http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

func (api *API) createNewUserHandler(c echo.Context) error {
	// Parse the request to get the email and password.
	req := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	// Call the CreateNewUser method from your Backend.
	user, err := api.backend.CreateNewUser(req.Email, req.Password)
	if err != nil {
		// Handle the error, e.g., return a JSON response with an error message.
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Return a JSON response with the created user.
	return c.JSON(http.StatusCreated, user)
}

func (api *API) resetPasswordHandler(c echo.Context) error {
	// Parse the request to get the email and newPassword.
	req := struct {
		Email       string `json:"email"`
		NewPassword string `json:"newPassword"`
	}{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	// Call the ResetUserPassword method from your Backend.
	err := api.backend.ResetUserPassword(req.Email, req.NewPassword)
	if err != nil {
		// Handle the error, e.g., return a JSON response with an error message.
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Return a success response.
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password reset successful.",
	})
}
