package kubeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ErrResponse error

type ErrDefault struct {
	Message  string `json:"message"`
	Code     int    `json:"code"`
	HttpCode int
}

func (e *ErrDefault) Error() string {
	return fmt.Sprintf("HTTP Code: %d, Code: %d, Message: %s", e.HttpCode, e.Code, e.Message)
}

func NewErrResponse(resp *http.Response) ErrResponse {
	errResponse := ErrDefault{
		HttpCode: resp.StatusCode,
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		errResponse.Message = err.Error()
		return &errResponse
	}

	//log.Printf("%s", bs)

	err = json.Unmarshal(bs, &errResponse)
	if err != nil {
		errResponse.Message = err.Error()
		return &errResponse
	}

	return &errResponse
}

type BadRequest struct {
	ErrDefault
	Extras struct {
		FailedFields []string `json:"failed_fields"`
	} `json:"extras"`
}

func (e *BadRequest) Error() string {
	return fmt.Sprintf("HTTP Code: %d, Code: %d, Message: %s, Failed Fields: %+v", e.HttpCode, e.Code, e.Message, e.Extras.FailedFields)
}

func NewErrBadRequest(resp *http.Response) ErrResponse {
	errResponse := BadRequest{
		ErrDefault: ErrDefault{
			HttpCode: resp.StatusCode,
		},
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		errResponse.Message = err.Error()
		return &errResponse
	}

	//log.Printf("%s", bs)

	err = json.Unmarshal(bs, &errResponse)
	if err != nil {
		errResponse.Message = err.Error()
		return &errResponse
	}

	return &errResponse
}
