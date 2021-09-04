//go:build embed_config
// +build embed_config

package config

import (
	_ "embed"
)

//go:embed config.json
var embedConfig []byte

func init() {
	embedRawConfig = embedConfig
}
