package fibergh

import (
	"cmp"
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/sse"
)

func GH[Req, Resp any](handlerFunc func(context.Context, *Req) (Resp, int, error)) fiber.Handler {
	return func(c fiber.Ctx) error {
		req := new(Req)

		err := c.Bind().All(req)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		resp, statusCode, err := handlerFunc(c, req)
		if err != nil {
			return c.Status(statusCode).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		err = c.Bind().RespHeader(resp)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		err = encodeHeader(c, resp)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		err = encodeCookie(c, resp)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if statusCode == fiber.StatusNoContent {
			return c.SendStatus(statusCode)
		}

		return c.Status(statusCode).JSON(resp)
	}
}

func encodeHeader(c fiber.Ctx, resp any) error {
	v := reflect.ValueOf(resp)

	if v.IsNil() {
		return nil
	}

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()

	for i := range v.NumField() {
		tag := t.Field(i).Tag

		headerTagValue := strings.TrimSpace(tag.Get("header"))

		if headerTagValue == "" {
			continue
		}

		c.Set(headerTagValue, v.Field(i).String())
	}

	return nil
}

var (
	rgxClearCookie = regexp.MustCompile(`^(.+),clear$`)
	rgxSetCookie   = regexp.MustCompile(`^(.+),([0-9hmsnu]+)$`)
)

func encodeCookie(c fiber.Ctx, resp any) error {
	v := reflect.ValueOf(resp)

	if v.IsNil() {
		return nil
	}

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()

	for i := range v.NumField() {
		tag := t.Field(i).Tag

		cookieTagValue := strings.TrimSpace(tag.Get("cookie"))

		if cookieTagValue == "" {
			continue
		}

		clearSm := rgxClearCookie.FindStringSubmatch(cookieTagValue)

		if len(clearSm) > 0 {
			c.ClearCookie(clearSm[1])

			continue
		}

		setSm := rgxSetCookie.FindStringSubmatch(cookieTagValue)

		if len(setSm) > 0 {
			duration, err := time.ParseDuration(setSm[2])
			if err != nil {
				return fmt.Errorf("time.ParseDuration: %w", err)
			}

			maxAge := 0
			maxAgeStr := tag.Get("cookieMaxAge")
			if maxAgeStr != "" {
				maxAge, err = strconv.Atoi(maxAgeStr)
				if err != nil {
					return fmt.Errorf("strconv.Atoi: %w", err)
				}
			}

			c.Cookie(&fiber.Cookie{
				Name:        setSm[1],
				Value:       v.Field(i).String(),
				Path:        cmp.Or(tag.Get("cookiePath"), "/"),
				Expires:     time.Now().Add(duration),
				SameSite:    cmp.Or(tag.Get("cookieSameSite"), "Lax"), // Lax (default), None, Strict
				Secure:      tag.Get("cookieSecure") == "true",
				HTTPOnly:    tag.Get("cookieHTTPOnly") == "true",
				Domain:      tag.Get("cookieDomain"), // default empty
				MaxAge:      maxAge,
				Partitioned: tag.Get("cookiePartitioned") == "true",
				SessionOnly: tag.Get("cookieSessionOnly") == "true",
			})
		}
	}

	return nil
}

func GHforSSE[Req, Data any](retry time.Duration, handlerFunc func(context.Context, *Req, func(name string, data Data) error) error) fiber.Handler {
	return sse.New(sse.Config{
		Retry: retry,
		Handler: func(c fiber.Ctx, stream *sse.Stream) error {
			req := new(Req)

			err := c.Bind().All(req)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": err.Error(),
				})
			}

			err = handlerFunc(c, req, func(name string, data Data) error {
				return stream.Event(sse.Event{
					ID:   uuid.NewV7().String(),
					Name: name,
					Data: data,
				})
			})
			if err != nil {
				return err
			}
			return nil
		},
	})
}
