

build:
	go build -o server ./cmd/app.go

build_embed_config:
	go build -o server -tags embed_config cmd/app.go
