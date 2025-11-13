package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
)

func errorMessage(c *fiber.Ctx, status int, err error) error {
	if errors.Is(err, storecmn.ErrNotFound) {
		status = fiber.StatusNotFound
	}

	return c.Status(status).JSON(fiber.Map{
		"error": err.Error(),
	})
}

func successMessage(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": message,
	})
}
