package metadata

import (
	"net/http"
	"net/url"
	"os"

	"github.com/rs/zerolog"
)

// NewContentManager returns a new RemoteContent object that
// implements the ContentManager interface.
func NewContentManager(serviceUrl string, client *http.Client, cacheEnabled bool) (*RemoteContent, error) {
	parsedUrl, err := url.Parse(serviceUrl)
	if err != nil {
		return nil, err
	}
	return &RemoteContent{
		ServiceUrl:                   parsedUrl,
		Client:                       client,
		Logger:                       zerolog.New(os.Stdout).With().Timestamp().Logger(),
		ContentConfigCache:           map[string]*ContentConfig{},
		ContentEncryptionConfigCache: map[string]*ContentEncryptionConfig{},
		CacheEnabled:                 cacheEnabled,
	}, nil
}

// ContentManager consisting of core methods to manage content metadata.
type ContentManager interface {
	GetConfig(string, *ContentConfig) (int, error)
	GetEncryptionConfig(string, string, *ContentEncryptionConfig, *string) (int, error)
}
