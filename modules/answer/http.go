package answer

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type Module struct {
	slackClient *slack.Client
}

func New(slackClient *slack.Client) *Module {
	return &Module{
		slackClient: slackClient,
	}
}

func (m *Module) Run() error {
	f := fiber.New()

	f.Use(recover.New())

	f.Post("/slack_events", func(c *fiber.Ctx) error {
		body := c.Body()

		var req http.Request
		err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true)
		if err != nil {
			c.Status(http.StatusServiceUnavailable)
			return err
		}

		sv, err := slack.NewSecretsVerifier(req.Header, "fab1308090a953a4380b6950178b74e2")
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
				// хак, иначе происходит пиздец
				var ts = ev.TimeStamp
				if ev.ThreadTimeStamp != "" {
					ts = ev.ThreadTimeStamp
				}

				fmt.Println(ev.Text)
				// ответить "иди нах"
				_, _, e := m.slackClient.PostMessage(
					ev.Channel,
					slack.MsgOptionTS(ts),
					slack.MsgOptionText("ПошЕЛ НАХУЙ", false),
				)
				if e != nil {
					log.Error().Err(err).Msg("send answer failed")
				}
			}
		}

		return nil
	})

	return f.Listen(":10001")
}
