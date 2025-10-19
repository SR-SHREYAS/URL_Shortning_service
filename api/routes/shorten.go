package routes

import (
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

	body := new(request)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}

	// implementing rate limiting
	r2 := database.CreateClient(1) // connect to redis client db 1 where rate limits are stored
	defer r2.Close()

	value, err := r2.Get(database.Ctx, c.IP()).Result() // get user ip to value
	if err == redis.Nil {
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else {
		valInt, _ := strconv.Atoi(value)
		if valInt <= 0 { // check on how many api requests are left
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result() // TTL is used to get the time to live of the key
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "rate limit exceeded", "rate_limit_reset": limit})
		}
	} // Note: A non-nil error other than redis.Nil is ignored here. Consider logging it.

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

	// connect to db 0 , where url are stored
	r := database.CreateClient(0)
	defer r.Close()

	//check if custom short is already in use
	value, _ = r.Get(database.Ctx, id).Result()
	if value != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "url custom short is already in use"})
	}

	// if expiry is not provided set it to 24 hours
	if body.Expiry == 0 {
		body.Expiry = 24
	}

	// set the url in the database with expiry in seconds (input is in hours)
	err = r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()
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

	// decrement the counter of requests limit
	r2.Decr(database.Ctx, c.IP())

	// get the number of requests remaining
	val, _ := r2.Get(database.Ctx, c.IP()).Result()
	response.XRateRemaining, _ = strconv.Atoi(val)

	// get the time to reset the counter
	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	response.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	// return the response , construct full short url
	response.CustomShort = os.Getenv("DOMAIN") + "/" + id

	// status 200
	return c.Status(fiber.StatusOK).JSON(response)
}
