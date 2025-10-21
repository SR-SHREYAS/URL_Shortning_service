package routes

import (
	"log"
	"os"
	"strconv"
	"time"
	"url_shortner_service/database"
	"url_shortner_service/helper"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// shape of the request expected from the user
type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"custom_short"`
	Expiry      time.Duration `json:"expiry"`
}

// shape of the response sent to the user
type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"custom_short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_remaining"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

// function to shorten the url
func ShortenURL(c *fiber.Ctx) error {

	ctx := c.Context()
	body := new(request)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}

	// --- Implementing Rate Limiting using the shared Redis client (Rdb1) ---
	val, err := database.Rdb1.Get(ctx, c.IP()).Result()
	if err == redis.Nil {
		// User's first request, set the API quota.
		// Use a pipeline to be slightly more efficient.
		pipe := database.Rdb1.Pipeline()
		pipe.Set(ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second)
		pipe.Exec(ctx)
	} else if err != nil {
		log.Printf("Error getting rate limit from Redis DB1: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database connection error"})
	} else {
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 { // check on how many api requests are left
			limit, _ := database.Rdb1.TTL(ctx, c.IP()).Result() // TTL is used to get the time to live of the key
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "rate limit exceeded", "rate_limit_reset": limit})
		}
	}

	// check if the input sent by user is valid url format
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid url"})
	}

	// check for domain error its not our own service domain
	if !helper.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "domain error"})
	}

	// enforce https , SSL , make sure url starts with http:// or https://
	body.URL = helper.EnforceHTTP(body.URL)

	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6] // customshort not provided so random 6 char uuid ID
	} else {
		id = body.CustomShort // use custom short
	}

	// Note: id is part after Domain Name , we create a custom string from logic and add our own domain name to it

	// --- Check if custom short is already in use using the shared client (Rdb0) ---
	// Use the efficient GET + redis.Nil check pattern.
	_, err = database.Rdb0.Get(ctx, id).Result()
	if err != redis.Nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "url custom short is already in use"})
	}

	// if expiry is not provided set it to 24 hours
	if body.Expiry == 0 {
		body.Expiry = 24
	}

	// set the url in the database with expiry in seconds (input is in hours)
	err = database.Rdb0.Set(ctx, id, body.URL, body.Expiry*3600*time.Second).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not shorten url"})
	}

	// response object are filled with value computed till now
	response := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	// Decrement the rate limit counter and get the new values.
	// Using a pipeline is efficient as it sends both commands in one round-trip.
	pipe := database.Rdb1.Pipeline()
	pipe.Decr(ctx, c.IP())
	pipe.TTL(ctx, c.IP())
	cmds, err := pipe.Exec(ctx)
	if err == nil {
		// Get remaining requests
		remaining, _ := database.Rdb1.Get(ctx, c.IP()).Result()
		response.XRateRemaining, _ = strconv.Atoi(remaining)

		// Get TTL from pipeline result
		if ttlCmd, ok := cmds[1].(*redis.DurationCmd); ok {
			response.XRateLimitReset = ttlCmd.Val() / time.Minute
		}
	}

	// return the response , construct full short url
	response.CustomShort = os.Getenv("DOMAIN") + "/" + id

	// status 200
	return c.Status(fiber.StatusOK).JSON(response)
}
