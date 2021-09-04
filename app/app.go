package app

import (
	"context"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/blessedvictim/frimon-bot/config"
	"github.com/blessedvictim/frimon-bot/model"
	"github.com/go-co-op/gocron"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

type App struct {
	cfg *config.Config

	slackClient *slack.Client

	httpClient *http.Client
}

func New(slackClient *slack.Client, httpClient *http.Client, cfg *config.Config) *App {
	return &App{
		cfg:         cfg,
		slackClient: slackClient,
		httpClient:  httpClient,
	}
}

func (a *App) Run() error {
	scheduler := gocron.NewScheduler(time.UTC)

	for _, job := range a.cfg.Jobs {
		_, err := scheduler.Cron(job.Cron).Do(a.executor, job)
		if err != nil {
			return errors.Wrapf(err, "cron job #%s failed", job.ID)
		}
	}

	scheduler.StartBlocking()

	return nil
}

func (a *App) executor(job model.Job) error {
	ctx := context.Background()

	i := rand.Intn(len(job.ContentList))
	content := job.ContentList[i]

	var err error
	switch content.Type {
	case model.ContentTypeImage:
		err = a.sendImage(ctx, job.SlackChannel, content)
	case model.ContentTypeFileLocal:
		err = a.sendFileLocal(ctx, job.SlackChannel, content)
	case model.ContentTypeFile:
		err = a.sendFile(ctx, job.SlackChannel, content)
	default:
		return errors.Errorf("undefined content type %s", content.Type)
	}

	if err != nil {
		log.Error().
			Err(err).
			Interface("job", job).
			Interface("content", content).
			Msg("executor failed")
	}

	return err
}

func (a *App) sendImage(ctx context.Context, channel string, content model.Content) error {
	var title string
	if content.Text != nil {
		title = *content.Text
	}

	_, _, _, err := a.slackClient.SendMessageContext(
		ctx,
		channel,
		slack.MsgOptionAsUser(true),
		slack.MsgOptionUsername("Friday here"),
		slack.MsgOptionBlocks(
			slack.NewImageBlock(
				content.Path,
				content.ID,
				"",
				slack.NewTextBlockObject(
					"plain_text", title, true, false),
			),
		),
	)
	if err != nil {
		return errors.Wrap(err, "send image message failed")
	}

	return nil
}

func (a *App) sendFile(ctx context.Context, chanel string, content model.Content) error {
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
		ThreadTimestamp: "",
	})
	if err != nil {
		return errors.Wrap(err, "send file failed")
	}

	return nil
}

func (a *App) sendFileLocal(ctx context.Context, channel string, content model.Content) error {
	file, err := a.getFileLocal(content.Path)
	if err != nil {
		return errors.Wrap(err, "get local file failed")
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
		Channels:        []string{channel},
		ThreadTimestamp: "",
	})
	if err != nil {
		return errors.Wrap(err, "send file failed")
	}

	return nil
}

// getFile load file and return request body io.ReadCloser, which must be closed by caller
func (a *App) getFileHTTP(ctx context.Context, path string) (io.ReadCloser, error) {
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

// getFile open file and return file io.ReadCloser, which must be closed by caller
func (a *App) getFileLocal(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "open file failed")
	}

	return f, nil
}
