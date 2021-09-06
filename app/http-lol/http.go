package http_lol

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func Run() error {
	f := fiber.New()

	f.Use(recover.New())

	f.Get("/slack_events", func(c *fiber.Ctx) error {
		body := c.Body()

		var req http.Request
		err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true)
		if err != nil {
			c.Status(http.StatusServiceUnavailable)
			return err
		}

		sv, err := slack.NewSecretsVerifier(req.Header, "")
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}
		if _, err := sv.Write(body); err != nil {
			return c.SendStatus(http.StatusInternalServerError)

		}
		if err := sv.Ensure(); err != nil {
			return c.SendStatus(http.StatusUnauthorized)
		}
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			return c.SendStatus(http.StatusInternalServerError)
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal(body, &r)
			if err != nil {
				return c.SendStatus(http.StatusInternalServerError)
			}
			c.Request().Header.Set("Content-Type", "text")
			_, _ = c.Write([]byte(r.Challenge))
		}
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				fmt.Println(ev.Text)
				// TODO
				//api.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello.", false))
			}
		}

		return nil
	})

	return f.Listen(":10001")
}
