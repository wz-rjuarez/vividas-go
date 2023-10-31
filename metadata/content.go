package metadata

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog"
)

type ContentEncryptionConfig struct {
	SessionBasedEncryptionPercentage int    `json:"sessionBasedEncryptionPercentage"`
	VivEncryptionPercentage          int    `json:"vivEncryptionPercentage"`
	ContentType                      string `json:"contentType"`
	ContentName                      string `json:"contentName"`
	ConvertToVod                     bool   `json:"convertToVod"`
	ChosenFrom                       string `json:"chosenFrom"`
	EncryptionPercentagesPerBitrates []struct {
		Quality              string `json:"quality"`
		EncryptionPercentage int    `json:"encryptionPercentage"`
	} `json:"encryptionPercentagesPerBitrates"`
	RawData string `json:"-"`
}

type ContentConfig struct {
	Uuid                             string    `json:"uuid"`
	PartnerUuid                      string    `json:"partnerUuid"`
	ContentName                      string    `json:"contentName"`
	ContentType                      string    `json:"contentType"`
	SessionBasedEncryptionPercentage int       `json:"sessionBasedEncryptionPercentage"`
	VivEncryptionPercentage          int       `json:"vivEncryptionPercentage"`
	Available                        bool      `json:"available"`
	ConvertToVod                     bool      `json:"convertToVod"`
	StorageType                      string    `json:"storageType"`
	CdnUrl                           string    `json:"cdnUrl"`
	Path                             string    `json:"path"`
	Status                           string    `json:"status"`
	CreatedAt                        time.Time `json:"createdAt"`
	UpdatedAt                        time.Time `json:"updatedAt"`
	DeletedAt                        time.Time `json:"deletedAt"`
}

// RemoteContent manages content metadata by making http requests to
// the encryption-metadata service. It implements the ContentManager
// interface.
type RemoteContent struct {
	ServiceUrl                   *url.URL
	Client                       *http.Client
	Logger                       zerolog.Logger
	ContentConfigCache           map[string]*ContentConfig
	ContentEncryptionConfigCache map[string]*ContentEncryptionConfig
	CacheEnabled                 bool
}

// GetConfig retrieves the metadata configuration belongs to a content id
// and stores the result in the value pointed to by v.
func (rc *RemoteContent) GetConfig(id string, v *ContentConfig) (int, error) {
	cacheKey := id
	if rc.CacheEnabled {
		val, ok := rc.ContentConfigCache[cacheKey]
		if ok {
			rc.Logger.Info().Msg("content config retrieved from cache")
			v = val
			rc.Logger.Debug().Msgf("content config:%+v", *v)
			return http.StatusOK, nil
		}
	}

	endpoint := rc.ServiceUrl.JoinPath("/contents/" + id).String()

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		rc.Logger.Error().Msg(err.Error())
		return http.StatusInternalServerError, err
	}

	rc.Logger.Debug().Msgf("requesting content config from %s", endpoint)
	resp, err := rc.Client.Do(req)
	if err != nil {
		rc.Logger.Error().Msg(err.Error())
		return http.StatusInternalServerError, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			rc.Logger.Error().Msg(err.Error())
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			rc.Logger.Error().Msg(err.Error())
			return http.StatusInternalServerError, err
		}
		msg := string(body)
		rc.Logger.Debug().Msgf("%d - %s", resp.StatusCode, msg)
		return resp.StatusCode, fmt.Errorf("%s", msg)
	}

	rc.Logger.Info().Msg("decoding content config...")
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		rc.Logger.Error().Msg(err.Error())
		return http.StatusInternalServerError, err
	}
	rc.Logger.Info().Msg("content config successfully decoded")
	rc.Logger.Debug().Msgf("content config:%+v", *v)

	if rc.CacheEnabled {
		rc.ContentConfigCache[cacheKey] = v
	}

	return resp.StatusCode, nil
}

// GetEncryptionConfig retrieves the metadata encryption configuration
// belongs to a content id and associated to a bitrate value.
// It stores the result in the value pointed to by v.
func (rc *RemoteContent) GetEncryptionConfig(id string, bitrate string, v *ContentEncryptionConfig) (int, error) {
	cacheKey := id + bitrate
	if rc.CacheEnabled {
		val, ok := rc.ContentEncryptionConfigCache[cacheKey]
		if ok {
			rc.Logger.Info().Msg("content encryption config retrieved from cache")
			v = val
			rc.Logger.Debug().Msgf("content encryption config:%+v", *v)
			return http.StatusOK, nil
		}
	}

	encryptionConfigUrl := rc.ServiceUrl.JoinPath("/contents/" + id + "/encryption-percentage")
	values := encryptionConfigUrl.Query()
	values.Add("bitrate", bitrate)
	encryptionConfigUrl.RawQuery = values.Encode()
	endpoint := encryptionConfigUrl.String()

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		rc.Logger.Error().Msg(err.Error())
		return http.StatusInternalServerError, err
	}

	rc.Logger.Debug().Msgf("requesting content encryption config from %s", endpoint)
	resp, err := rc.Client.Do(req)
	if err != nil {
		rc.Logger.Error().Msg(err.Error())
		return http.StatusInternalServerError, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			rc.Logger.Error().Msg(err.Error())
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		rc.Logger.Error().Msg(err.Error())
		return http.StatusInternalServerError, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := string(body)
		rc.Logger.Debug().Msgf("%d - %s", resp.StatusCode, msg)
		return resp.StatusCode, fmt.Errorf("%s", msg)
	}

	rc.Logger.Info().Msg("decoding content encryption config...")
	err = json.Unmarshal(body, &v)
	if err != nil {
		rc.Logger.Error().Msg(err.Error())
		return http.StatusInternalServerError, err
	}
	v.RawData = string(body)
	rc.Logger.Info().Msg("content encryption config successfully decoded")
	rc.Logger.Debug().Msgf("content encryption config:%+v", *v)

	if rc.CacheEnabled {
		rc.ContentEncryptionConfigCache[cacheKey] = v
	}

	return resp.StatusCode, nil
}
