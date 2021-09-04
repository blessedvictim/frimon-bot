package model

import "time"

type Job struct {
	ID           string    `mapstructure:"id"`
	Cron         string    `mapstructure:"cron"`
	SlackChannel string    `mapstructure:"slack_channel"`
	ContentList  []Content `mapstructure:"content_list"`
}

type Content struct {
	ID   string      `mapstructure:"id"`
	Type ContentType `mapstructure:"type"`
	Text *string     `mapstructure:"text"`
	Path string      `mapstructure:"path"`
}

type ContentType string

const (
	ContentTypeImage     ContentType = "image"
	ContentTypeFile      ContentType = "file"
	ContentTypeFileLocal ContentType = "file-local"
)

type JobLog struct {
	IDJob       int       `mapstructure:"id_job"`
	PerformedAt time.Time `mapstructure:"performed_at"`
	IDContent   int       `mapstructure:"id_content"`
}
