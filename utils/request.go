package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type APIResponse[T any] struct {
	Result   T        `json:"result"`
	Success  bool     `json:"success"`
	Messages []string `json:"messages"`
	Errors   []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func DoRequest[T any](method string, url string, payload map[string]any, apiToken string) (*APIResponse[T], error) {
	var reqbody io.Reader
	if payload != nil {
		jsonString, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		reqbody = bytes.NewReader(jsonString)
	}
	req, err := http.NewRequest(method, url, reqbody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var apiRes APIResponse[T]
	if err := json.Unmarshal(body, &apiRes); err != nil {
		return nil, err
	}

	return &apiRes, nil
}
