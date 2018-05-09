package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/pkg/errors"
)

const (
	loginURI                = "/zstack/v1/accounts/login"
	logoutURI               = "/zstack/v1/accounts/sessions/{uuid}"
	defaultQueryAsyncPeriod = 1 * time.Second
)

type ResourceBase struct {
	UUID       string `json:"uuid,omitempty"`
	CreateDate string `json:"createDate,omitempty"`
	LastOpDate string `json:"lastOpDate,omitempty"`
}

type Tags struct {
	SystemTags []string `json:"systemTags,omitempty"`
	UserTags   []string `json:"userTags,omitempty"`
}

type LoginRequest struct {
	LoginContent map[string]string `json:"logInByAccount,omitempty"`
	Tags         `json:",inline"`
}

type LoginResponse struct {
	Inventory struct {
		UUID        string `json:"uuid,omitempty"`
		AccountUUID string `json:"accountUuid,omitempty"`
		UserUUID    string `json:"userUuid,omitempty"`
		ExpiredDate string `json:"expiredDate,omitempty"`
		CreateDate  string `json:"createDate,omitempty"`
	} `json:"inventory,omitempty"`
}

type ErrorResponse struct {
	Error	Error `json:"error,omitempty"`
}

type ZStack503Error struct {
	Error *Error `json:"error,omitempty"`
}

type Error struct {
	Code        string            `json:"code,omitempty"`
	Description string            `json:"description,omitempty"`
	Details     string            `json:"details,omitempty"`
	Elaboration string            `json:"elaboration,omitempty"`
	Opaque      map[string]string `json:"opaque,omitemtpy"`
	Cause       *Error            `json:"cause,omitempty"`
}

type AsyncResponse struct {
	Location string `json:"location"`
	client   *Client
}

func (Error *Error) WrapError() error {
	wrapError := fmt.Errorf("code:%s,detail:%s,description:%s",
		Error.Code,
		Error.Details,
		Error.Description,
	)
	for i := 0; i < 10; i++ {
		if Error.Cause == nil {
			break
		}
		wrapError = errors.Wrap(
			wrapError,
			fmt.Sprintf("code:%s,detail:%s,description:%s",
				Error.Cause.Code,
				Error.Cause.Details,
				Error.Cause.Description,
			),
		)
	}
	return wrapError
}

func GetAsyncResponse(c *Client, resp *http.Response) (*AsyncResponse, error) {
	if resp.StatusCode != 202 {
		return nil, errors.New("can't parse a non-async response")
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	rtn := AsyncResponse{}
	if err = json.Unmarshal(responseBody, &rtn); err != nil {
		return nil, err
	}
	rtn.client = c
	return &rtn, nil
}

func (async *AsyncResponse) QueryRealResponse(i interface{}, timeout time.Duration) error {
	timeouter := time.After(timeout)
	ticker := time.NewTicker(defaultQueryAsyncPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-timeouter:
			return fmt.Errorf("querying location: %s timeout", async.Location)
		case <-ticker.C:
			resp, err := async.QueryLocation()
			if err != nil {
				return err
			}
			switch resp.StatusCode {
			case 202:
				//If job still running, ZStack will return code 202 continely.
				resp.Body.Close()
				continue
			case 200:
				//Job success
				responseBody, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				logrus.Info(string(responseBody))
				if string(responseBody) == "" {
					return nil
				}
				return json.Unmarshal(responseBody, i)
			case 404:
				//Location is no longer available
				resp.Body.Close()
				return fmt.Errorf("location %s is no longer available", async.Location)
			case 503:
				responseBody, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					logrus.WithError(err).Error("the ZStack returns a code 503")
					return err
				}
				zstack503Error := ZStack503Error{}
				if err = json.Unmarshal(responseBody, &zstack503Error); err != nil {
					logrus.WithError(err).Error("the ZStack returns a code 503")
					return err
				}
				return zstack503Error.Error.WrapError()
			default:
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					logrus.WithError(err).Errorf("get status code %d and get data error", resp.StatusCode)
				}
				return fmt.Errorf("status code %d and error response %s", resp.StatusCode, string(body))
			}
		}
	}
}

func (async *AsyncResponse) QueryLocation() (*http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, async.Location, nil)
	if err != nil {
		return nil, err
	}
	return async.client.httpClient.Do(request)
}
