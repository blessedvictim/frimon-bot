package answer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/blessedvictim/frimon-bot/model"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type Module struct {
	slackClient *slack.Client
	httpClient  *http.Client
}

func New(slackClient *slack.Client) *Module {
	return &Module{
		slackClient: slackClient,
		httpClient:  http.DefaultClient,
	}
}

var answers = []string{
	"Я новичок в мире slack, но уже понял что ты самый невзрачный персонаж здесь, не приставай, не стоит вскрывать эту тему...",
	"Я не хочу общаться с бичами",
	"Я тебя прошу, не беспокой меня",
	"<msg>https://i.imgur.com/xZy0zkA.png", // угомонись чОРт
	"<msg>https://i.imgur.com/jw2D4Aw.png", // моника
}

func (m *Module) Run() error {
	f := fiber.New()

	f.Use(recover.New())

	var userMentionsCount = make(map[string]int)

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
		eventsAPIEvent, err := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
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
		BreakLabel:
			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				// хак, иначе происходит пиздец
				var ts = ev.TimeStamp
				if ev.ThreadTimeStamp != "" {
					ts = ev.ThreadTimeStamp
				}

				count := userMentionsCount[ev.User]
				userMentionsCount[ev.User] = count + 1

				answerMsgOptions := make([]slack.MsgOption, 0)

				// дикий костыль
				l := len(answers)
				switch {
				case l == 5 && count == 0:
					answerMsgOptions = append(answerMsgOptions, slack.MsgOptionText(answers[0], false))
					answers = answers[1:]
				case count <= l:
					content := answers[count-1]
					if strings.HasPrefix(content, "<msg>") {
						e := m.sendFile(
							context.Background(),
							ev.Channel,
							model.Content{
								ID:   content,
								Type: model.ContentTypeImage,
								Text: nil,
								Path: strings.ReplaceAll(content, "<msg>", ""),
							},
							ts,
						)
						if e != nil {
							log.Error().Err(err).Msg("send answer failed")
						}
						break BreakLabel
					} else {
						answerMsgOptions = append(answerMsgOptions, slack.MsgOptionText(answers[count-1], false))
					}
				default:
					break BreakLabel
				}

				// TODO cleanup
				fmt.Println(ev.Text)

				_, _, e := m.slackClient.PostMessage(
					ev.Channel,
					append([]slack.MsgOption{slack.MsgOptionTS(ts)}, answerMsgOptions...)...,
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

func (a *Module) sendFile(ctx context.Context, chanel string, content model.Content, ts string) error {
	file, err := a.getFileHTTP(ctx, content.Path)
	if err != nil {
		return errors.Wrap(err, "get file http failed")
	}

	defer func() {
		_ = file.Close()
	}()

	var title string
	if content.Text != nil {
		title = *content.Text
	}

	_, err = a.slackClient.UploadFileContext(ctx, slack.FileUploadParameters{
		File:            "",
		Content:         "",
		Reader:          file,
		Filetype:        "",
		Filename:        content.ID,
		Title:           title,
		InitialComment:  "",
		Channels:        []string{chanel},
		ThreadTimestamp: ts,
	})
	if err != nil {
		return errors.Wrap(err, "send file failed")
	}

	return nil
}

// getFile load file and return request body io.ReadCloser, which must be closed by caller
func (a *Module) getFileHTTP(ctx context.Context, path string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "make req failed")
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do req failed")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("request failed with code #%d", resp.StatusCode)
	}

	return resp.Body, nil
}
