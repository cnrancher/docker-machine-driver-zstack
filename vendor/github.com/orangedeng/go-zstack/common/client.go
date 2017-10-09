package common

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type Client struct {
	accountName    string
	password       string
	serverEndpoint string
	sessionID      string
	httpClient     *http.Client
}

func (client *Client) Init(AccountName, Password, ServerEndpoint string) error {
	hsha512 := sha512.New()
	io.WriteString(hsha512, Password)
	client.accountName = AccountName
	client.password = fmt.Sprintf("%x", hsha512.Sum(nil))
	client.serverEndpoint = ServerEndpoint
	client.httpClient = &http.Client{}
	return client.login()
}

func (client *Client) login() error {
	login := LoginRequest{
		LoginContent: map[string]string{
			"password":    client.password,
			"accountName": client.accountName,
		},
		Tags: Tags{
			SystemTags: []string{},
			UserTags:   []string{},
		},
	}
	requestBody, _ := json.Marshal(login)
	resp, err := client.CreateRequestWithURI(http.MethodPost, loginURI, requestBody)
	if err != nil {
		return errors.Wrap(err, "Get error while login request")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Get error while getting data from login response")
	}

	loginResponse := LoginResponse{}
	if err := json.Unmarshal(respBody, &loginResponse); err != nil {
		return errors.Wrap(err, "Get error while decoding login response")
	}
	if int(resp.StatusCode/100) != 2 {
		return fmt.Errorf("return status: %d,not equal to 2xx", resp.StatusCode)
	}
	if loginResponse.Error != nil {
		return errors.New(loginResponse.Error.Description + " " + loginResponse.Error.Details)
	}

	client.sessionID = loginResponse.Inventory.UUID
	return nil
}

func (client *Client) Cleanup() error {
	if client.sessionID == "" {
		return nil
	}
	return client.deleteSessionID()

}

func (client *Client) deleteSessionID() error {
	URI := strings.Replace(logoutURI, "{uuid}", client.sessionID, -1)
	resp, err := client.CreateRequestWithURI(http.MethodDelete, URI, nil)
	if err != nil {
		return errors.Wrap(err, "Get error while logout request")
	}
	if resp.StatusCode != 200 {
		return errors.New("Delete session id request does not get 200 response code")
	}
	return nil
}

func (client *Client) CreateRequestWithURI(method, uri string, body []byte) (*http.Response, error) {
	urlPath := client.serverEndpoint + uri
	httpRequest, err := http.NewRequest(method, urlPath, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Add("Authorization", "OAuth "+client.sessionID)
	resp, err := client.httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
