# fibergh

Type-safe HTTP handlers for [Fiber](https://gofiber.io/) using Go generics. Bind requests, encode responses, and handle errors with zero boilerplate.

## What is a Generic Handler?

A generic handler is a typed wrapper that removes boilerplate from your handlers. But the real benefit goes beyond convenience.

**Zero third-party dependencies in your handlers.** Your handler function receives a typed request struct and returns a typed response struct — no Fiber, no `fiber.Ctx`, no framework-specific code. This means your business logic is decoupled from the router.

```go
// This handler has ZERO framework imports
func CreateUserHandler(ctx context.Context, req *CreateUserReq) (*CreateUserResp, int, error) {
    // pure business logic
}
```

**Router-agnostic.** Since your handlers don't depend on Fiber, you could switch to chi, echo, or any other router without rewriting handler code. Router changes are rare in real projects, but this flexibility means your core logic is never held hostage by a framework choice.

**Readable request/response handling.** The most common boilerplate in web apps is reading requests and writing responses. Struct tags make this declarative and self-documenting — you see exactly what comes in and what goes out at a glance, without digging through handler code.

```go
// You can immediately see: name from body, age from query, token from cookie
type Req struct {
    Name  string `json:"name"`
    Age   int    `query:"age"`
    Token string `cookie:"token"`
}
```

## Install

```bash
go get github.com/ksckaan1/fibergh
```

## Usage

### `GH` — Generic HTTP Handler

`GH[Req, Resp]` wraps a typed handler function into a `fiber.Handler`. It automatically:

- Binds request body, query params, and path params into your `Req` struct
- Encodes response headers from `header` struct tags
- Sets/clears cookies from `cookie` struct tags
- Returns JSON responses with proper status codes
- Converts errors into `{"error": "..."}` JSON responses

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/ksckaan1/fibergh"
)

type CreateUserReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserResp struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Token string `cookie:"token,24h" json:"-"`
}

func CreateUserHandler(ctx context.Context, req *CreateUserReq) (*CreateUserResp, int, error) {
	// Your business logic here
	return &CreateUserResp{
		ID:    "abc-123",
		Name:  req.Name,
		Token: "secret-token",
	}, http.StatusCreated, nil
}

func main() {
	app := fiber.New()

	app.Post("/users", fibergh.GH(CreateUserHandler))

	log.Fatal(app.Listen(":3000"))
}
```

### Struct Tags

**Request binding** is powered by Fiber v3's [Bind](https://docs.gofiber.io/api/bind/) feature. `GH` calls `c.Bind().All(req)` under the hood, which binds data from URI params, request body, query params, headers, and cookies in that precedence order.

```go
type Req struct {
	Name      string `json:"name"`         // body
	Age       int    `query:"age"`         // query param
	UserID    string `header:"X-User-ID"`  // request header
	PostID    string `uri:"id"`            // path param
	SessionID string `cookie:"session_id"` // cookie
}
```

**Response encoding** is implemented by fibergh (not Fiber). It uses `header` and `cookie` struct tags on the response struct to set response headers and cookies.

```go
type Resp struct {
	Name      string `json:"name"`
	AuthToken string `cookie:"auth_token,1h" json:"-"`     // sets cookie, omits from JSON
	Session   string `cookie:"session,clear" json:"-"`     // clears cookie
	CustomID  string `header:"X-Custom-ID" json:"-"`       // sets response header
}
```

### Cookie Tag Options

The `cookie` tag format is `cookie:"<name>,<duration>"` or `cookie:"<name>,clear"`.

Additional options are set via separate struct tags:

| Tag                 | Description        | Default |
| ------------------- | ------------------ | ------- |
| `cookiePath`        | Cookie path        | `/`     |
| `cookieSameSite`    | SameSite attribute | `Lax`   |
| `cookieSecure`      | Secure flag        | `false` |
| `cookieHTTPOnly`    | HttpOnly flag      | `false` |
| `cookieDomain`      | Domain attribute   | (empty) |
| `cookieMaxAge`      | Max-Age in seconds | `0`     |
| `cookiePartitioned` | Partitioned flag   | `false` |
| `cookieSessionOnly` | Session-only flag  | `false` |

```go
type Resp struct {
	Token string `cookie:"token,24h" cookiePath:"/" cookieSecure:"true" cookieHTTPOnly:"true" json:"-"`
}
```

### Default Values

Set default values using the `default` struct tag. Supported types: `bool`, `int`, `uint`, `float`, `string`, slices (use `|` as separator), and pointers to these types.

```go
type Req struct {
	Name     string   `query:"name,default:john"`
	Age      int      `query:"age,default:18"`
	Products []string `query:"products,default:shoe|hat"`
}
```

### Validation

Add a struct validator to your Fiber config, then use the `validate` struct tag:

```go
import "github.com/go-playground/validator/v10"

app := fiber.New(fiber.Config{
	StructValidator: &structValidator{validate: validator.New()},
})

type Req struct {
	Name  string `json:"name" validate:"required"`
	Age   int    `json:"age" validate:"gte=0,lte=150"`
}
```

> For full details on request binding, default values, validation, and more, see the [Fiber Bind docs](https://docs.gofiber.io/api/bind/).

### `GHforSSE` — Server-Sent Events

`GHforSSE[Req, Data]` provides type-safe SSE with request binding and named events. It accepts a retry duration and a handler that receives the parsed request and a `send` function.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/ksckaan1/fibergh"
)

type ChatReq struct {
	RoomID string `uri:"id"`
}

type ChatEvent struct {
	Message string `json:"message"`
}

func chatHandler(ctx context.Context, req *ChatReq, send func(name string, data ChatEvent) error) error {
	for i := 0; i < 10; i++ {
		if err := send("message", ChatEvent{Message: fmt.Sprintf("hello %d", i)}); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
	return nil
}

func main() {
	app := fiber.New()

	app.Post("/stream/:id", fibergh.GHforSSE(5*time.Second, chatHandler))

	log.Fatal(app.Listen(":3000"))
}
```

## License

[MIT](LICENSE)
