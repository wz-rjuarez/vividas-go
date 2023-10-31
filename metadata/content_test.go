package metadata

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func TestGetConfig(t *testing.T) {
	getContentConfig := func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		contentId := vars["contentId"]

		_, err := uuid.Parse(contentId)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Value is not a valid UUID V4 string"))
			return
		}

		if (contentId != "f0121a13-8f2a-4dac-ab07-b49e10aeefcf") && (contentId != "9c02fc65-e782-4f85-af92-a3134e028515") {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(fmt.Sprintf("Content with UUID %s doesn't exist", contentId)))
			return
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`{
			"uuid":"f0121a13-8f2a-4dac-ab07-b49e10aeefcf",
			"partnerUuid":"3db5dabc-90e5-42fe-a286-a8eb720d9ee5",
			"contentName":"Sintel VOD Dash encrypted",
			"contentType":"video-on-demand",
			"sessionBasedEncryptionPercentage":20,
			"vivEncryptionPercentage":20,
			"available":false,
			"convertToVod":false,
			"storageType":"s3",
			"cdnUrl":"",
			"path":"sintel_dash",
			"status":"CREATED"
			}`))
	}

	data := []struct {
		name         string
		id           string
		expectedCode int
	}{
		{
			"success",
			"f0121a13-8f2a-4dac-ab07-b49e10aeefcf",
			http.StatusOK,
		},
		{
			"id is empty",
			"",
			http.StatusNotFound,
		},
		{
			"id doesn't exist",
			"d5583a9c-f4e3-4ca5-88cd-8403f50b4961",
			http.StatusBadRequest,
		},
		{
			"id missing last char",
			"f0121a13-8f2a-4dac-ab07-b49e10aeefc",
			http.StatusBadRequest,
		},
		{
			"id invalid first char",
			"W0121a13-8f2a-4dac-ab07-b49e10aeefcf",
			http.StatusBadRequest,
		},
		{
			"invalid id",
			"Mmm",
			http.StatusBadRequest,
		},
		{
			"success retrieving the same id (cache enabled)",
			"f0121a13-8f2a-4dac-ab07-b49e10aeefcf",
			http.StatusOK,
		},
		{
			"success by using a different id",
			"9c02fc65-e782-4f85-af92-a3134e028515",
			http.StatusOK,
		},
	}

	mux := mux.NewRouter()
	mux.HandleFunc("/contents/{contentId}", getContentConfig)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	contentManager, err := NewContentManager(ts.URL, ts.Client(), true)
	if err != nil {
		t.Error(err)
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			var contentConfig ContentConfig
			code, _ := contentManager.GetConfig(d.id, &contentConfig)
			if d.expectedCode != code {
				t.Errorf("expected code '%d', got '%d'", d.expectedCode, code)
			}
		})
	}
}

func TestGetEncryptionConfig(t *testing.T) {
	getContentEncryptionConfig := func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)

		contentId := vars["contentId"]
		_, err := uuid.Parse(contentId)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Value is not a valid UUID V4 string"))
			return
		}
		if (contentId != "f0121a13-8f2a-4dac-ab07-b49e10aeefcf") && (contentId != "9c02fc65-e782-4f85-af92-a3134e028515") {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(fmt.Sprintf("Content with UUID %s doesn't exist", contentId)))
			return
		}

		bitrate := req.URL.Query().Get("bitrate")
		isValidBitrate := true
		if bitrate == "" {
			isValidBitrate = false
		}
		if num, err := strconv.Atoi(bitrate); err != nil {
			isValidBitrate = false
		} else if num <= 0 {
			isValidBitrate = false
		}
		if !isValidBitrate {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Bitrates list is not compliant"))
			return
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`{
			"sessionBasedEncryptionPercentage":20,
			"vivEncryptionPercentage":20,
			"contentType":"video-on-demand",
			"contentName":"",
			"convertToVod":false,
			"chosenFrom":"ENCRYPTION_PERCENTAGE_TITLE",
			"encryptionPercentagesPerBitrates":[
				{
					"quality":"1080",
					"encryptionPercentage":50
				},
				{
					"quality":"720",
					"encryptionPercentage":40
				},
				{
					"quality":"480",
					"encryptionPercentage":30
				}
			]}`))
	}

	data := []struct {
		name         string
		id           string
		bitrate      string
		expectedCode int
	}{
		{
			"success with valid id",
			"f0121a13-8f2a-4dac-ab07-b49e10aeefcf",
			"1080",
			http.StatusOK,
		},
		{
			"id is empty",
			"",
			"1080",
			http.StatusNotFound,
		},
		{
			"id doesn't exist",
			"d5583a9c-f4e3-4ca5-88cd-8403f50b4961",
			"1080",
			http.StatusBadRequest,
		},
		{
			"invalid id",
			"Mmm",
			"1080",
			http.StatusBadRequest,
		},
		{
			"bitrate is empty",
			"f0121a13-8f2a-4dac-ab07-b49e10aeefcf",
			"",
			http.StatusBadRequest,
		},
		{
			"bitrate is zero",
			"f0121a13-8f2a-4dac-ab07-b49e10aeefcf",
			"0",
			http.StatusBadRequest,
		},
		{
			"bitrate is less than zero",
			"f0121a13-8f2a-4dac-ab07-b49e10aeefcf",
			"-30",
			http.StatusBadRequest,
		},
		{
			"success retrieving the same id and bitrate (cache enabled)",
			"f0121a13-8f2a-4dac-ab07-b49e10aeefcf",
			"1080",
			http.StatusOK,
		},
		{
			"success by using a different id",
			"9c02fc65-e782-4f85-af92-a3134e028515",
			"1080",
			http.StatusOK,
		},
	}

	mux := mux.NewRouter()
	mux.HandleFunc("/contents/{contentId}/encryption-percentage", getContentEncryptionConfig)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	contentManager, err := NewContentManager(ts.URL, ts.Client(), true)
	if err != nil {
		t.Error(err)
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			var contentEncryptionConfig ContentEncryptionConfig
			code, _ := contentManager.GetEncryptionConfig(d.id, d.bitrate, &contentEncryptionConfig)
			if d.expectedCode != code {
				t.Errorf("expected code '%d', got '%d'", d.expectedCode, code)
			}
		})
	}
}
