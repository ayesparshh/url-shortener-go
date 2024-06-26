package routes

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/ayesparshh/url-shortner-go/database"
	"github.com/ayesparshh/url-shortner-go/helpers"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type request struct {
	URL					string		     `json:"url"`
	CustomShortURL		string			 `json:"customshorturl"`
	Expiry				time.Duration	 `json:"expiry"`
}

type response struct {

	URL    				string			 `json:"url"`
	CustomShortURL      string			 `json:"customshorturl"`
	Expiry              time.Duration	 `json:"expiry"`
	XRateRemaining      int				 `json:"xrateremaining"`
	XRateLimitReset      time.Duration	 `json:"xratelimitreset"`

}

func ShortenURL(c *fiber.Ctx) error {

	body := new(request)

	if err := c.BodyParser(body); err != nil {
		fmt.Println("Error parsing body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// implement rate limiting
	r2 :=database.CreateClient(1)
	defer r2.Close()

	_,err := r2.Get(database.Ctx, c.IP()).Result()

	if err == redis.Nil {
		_ = r2.Set(database.Ctx, c.IP(),os.Getenv("Apilimit"),30*60*time.Second).Err()
	}else{
		val , _ := r2.Get(database.Ctx, c.IP()).Result()
		valInt, _ := strconv.Atoi(val)

		if valInt <= 0 {
			limit,_ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "rate limit exceeded", "limit reset in": limit/time.Nanosecond/time.Minute})
		}
	}
	//check is url sent by user is valid
	if !govalidator.IsURL(body.URL) {	
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid url"})
	}
	//check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "domain error 🤬🤬"})
	}
	//enfore hheps,ssl
	body.URL = helpers.EnforceHTTP(body.URL)

	var id string
	if body.CustomShortURL == "" {
		id = uuid.New().String()[:6]
	}else {
		id = body.CustomShortURL
	}	

	r := database.CreateClient(0)
	defer r.Close()

	val , _ := r.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "custom url already exists"})
	}

	if body.Expiry == 0 {
		body.Expiry = 24 * time.Hour
	}

	err = r.Set(database.Ctx, id, body.URL, body.Expiry).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error creating custom url"})
	}

	resp := response{
		URL:             body.URL,
		CustomShortURL:  "",
		Expiry: 		 body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}
	r2.Decr(database.Ctx, c.IP())
	
	val , _ = r2.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl ,_ := r2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttl/time.Nanosecond/time.Minute

	resp.CustomShortURL = os.Getenv("localhost:3000") + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)
}
