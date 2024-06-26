package routes

import (
	"github.com/ayesparshh/url-shortner-go/database"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

func ResolveURL(c *fiber.Ctx) error {

	url := c.Params("url")

    rdb := database.CreateClient(0)
	defer rdb.Close()

	value ,err := rdb.Get(database.Ctx, url).Result()

	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short url not found"})
    }else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot resolve short url"})
    }

	redir := database.CreateClient(1)
	defer redir.Close()

	_ = redir.Incr(database.Ctx, "counter")

	return c.Redirect(value , 301)

}
