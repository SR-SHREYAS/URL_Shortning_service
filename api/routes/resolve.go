package routes

import (
	"url_shortner_service/database"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

func ResolveURL(c *fiber.Ctx) error {
	url := c.Params("url") // input url from user to be resolved

	r := database.CreateClient(0) // connected to redis client db 0 where url is stored
	defer r.Close()

	value, err := r.Get(database.Ctx, url).Result() // fetch url with key of input url
	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short not found in the database"})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not resolve short url"})
	}

	// incrementing the counter to track how many times the short url is visited
	rInr := database.CreateClient(1)
	defer rInr.Close()

	_ = rInr.Incr(database.Ctx, "counter")

	// redirect to the original url
	return c.Redirect(value, fiber.StatusMovedPermanently)

}
