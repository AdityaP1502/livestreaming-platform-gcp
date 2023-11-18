package json

import (
	"encoding/json"
	"io"

	"github.com/AdityaP1502/livestreaming-platform-gcp/api/go/base"
)

func CreateJSONResponse(responseStatus string, responseMessage string, data base.DataField) ([]byte, error) {
	response := base.Response{
		Status:  responseStatus,
		Message: responseMessage,
		Data:    data,
	}

	jsonResponse, err := json.Marshal(response)

	if err != nil {
		return nil, err
	}

	return jsonResponse, nil
}

func DecodeJSONBody(r io.Reader, v interface{}) error {
	decoder := json.NewDecoder(r)

	var err error

	for {
		err = decoder.Decode(&v)
		if err != nil {

			if err == io.EOF {
				err = nil
			}

			break
		}
	}

	return err
}
