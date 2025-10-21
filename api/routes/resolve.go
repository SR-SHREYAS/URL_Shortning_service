package routes

import (
	"fmt"
	"log"
	"url_shortner_service/database"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

func ResolveURL(c *fiber.Ctx) error {
	url := c.Params("url") // input url from user to be resolved
	ctx := c.Context()     // Use the request's context for cancellation handling

	// Use the shared Redis client instance from the database package.
	// This client is created once when the app starts.
	value, err := database.Rdb0.Get(ctx, url).Result() // fetch url with key of input url
	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short not found in the database"})
	} else if err != nil {
		log.Printf("Error resolving URL from Redis DB0: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not resolve short url"})
	}

	// Increment the counter for this specific URL to track its usage.
	// We create a unique key for each URL's counter.
	counterKey := fmt.Sprintf("counter:%s", url)
	if err := database.Rdb1.Incr(ctx, counterKey).Err(); err != nil {
		// Log the error if the counter fails to increment, but don't block the redirect.
		log.Printf("Error incrementing counter in Redis DB1: %v", err)
	}

	// redirect to the original url
	return c.Redirect(value, fiber.StatusMovedPermanently)
}
