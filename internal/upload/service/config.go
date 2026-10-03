package upload_service

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type UploadConfig struct {
	// Dir is where uploaded files are stored on the local disk.
	Dir string `envconfig:"DIR" default:"./data/uploads"`
	// PublicBaseUrl is the origin put in front of /uploads/{name} in the URLs returned to clients,
	// e.g. "https://api.tailverse.app". When empty, the scheme and Host of the upload request are used.
	PublicBaseUrl string `envconfig:"PUBLIC_BASE_URL"`
}

func NewUploadConfigMust() UploadConfig {
	var config UploadConfig
	if err := envconfig.Process("UPLOADS", &config); err != nil {
		panic(fmt.Errorf("process uploads config: %w", err))
	}
	return config
}
