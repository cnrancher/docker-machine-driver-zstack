package instance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/orangedeng/go-zstack/common"
)

type Client struct {
	common.Client
}

func (c *Client) CreateInstance(req CreateRequest) (*common.AsyncResponse, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.CreateRequestWithURI(http.MethodPost, createInstanceURI, requestBody)
	if err != nil {
		return nil, err
	}
	return common.GetAsyncResponse(&c.Client, resp)
}

func (c *Client) DeleteInstance(UUID string) (*common.AsyncResponse, error) {
	realURI := strings.Replace(deleteInstanceURI, "{uuid}", UUID, -1)
	resp, err := c.Client.CreateRequestWithURI(http.MethodDelete, realURI, nil)
	if err != nil {
		return nil, err
	}
	return common.GetAsyncResponse(&c.Client, resp)
}

func (c *Client) ExpungeInstance(UUID string) (*common.AsyncResponse, error) {
	tmp := ExpungeInstanceRequest{
		ExpungeVMInstance: map[string]string{},
	}
	requestBody, err := json.Marshal(tmp)

	if err != nil {
		return nil, err
	}
	realURI := strings.Replace(operateInstanceURI, "{uuid}", UUID, -1)

	resp, err := c.Client.CreateRequestWithURI(http.MethodPut, realURI, requestBody)
	if err != nil {
		return nil, err
	}
	return common.GetAsyncResponse(&c.Client, resp)
}

func (c *Client) QueryInstance(UUID string) (*VMInstanceInventory, error) {
	realURI := strings.Replace(queryInstanceURI, "{uuid}", UUID, -1)
	resp, err := c.Client.CreateRequestWithURI(http.MethodGet, realURI, nil)
	if err != nil {
		return nil, err
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseStruct := QueryInstanceResponse{}
	if err = json.Unmarshal(responseBody, &responseStruct); err != nil {
		logrus.Warnf("Unmarshaling response when Querying instance. Error: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		if responseStruct.Error != nil {
			return nil, responseStruct.Error.WrapError()
		}
		return nil, fmt.Errorf("status code %d,Error massage %s", resp.StatusCode, string(responseBody))
	}
	if len(responseStruct.Inventories) == 0 {
		return nil, fmt.Errorf("can't get any instance informations, expect one")
	}
	return responseStruct.Inventories[0], nil
}

func (c *Client) QueryInstances() ([]*VMInstanceInventory, error) {
	resp, err := c.Client.CreateRequestWithURI(http.MethodGet, queryInstancesURI, nil)
	if err != nil {
		return nil, err
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseStruct := QueryInstanceResponse{}
	if err = json.Unmarshal(responseBody, &responseStruct); err != nil {
		logrus.Warnf("Unmarshaling response when Querying instance. Error: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		if responseStruct.Error != nil {
			return nil, responseStruct.Error.WrapError()
		}
		return nil, fmt.Errorf("status code %d,Error massage %s", resp.StatusCode, string(responseBody))
	}
	return responseStruct.Inventories, nil
}

func (c *Client) StartInstance(UUID string) (*common.AsyncResponse, error) {
	requestStruct := StartInstanceRequest{
		StartVMInstance: map[string]string{},
	}
	requestBody, err := json.Marshal(requestStruct)
	if err != nil {
		return nil, err
	}

	realURI := strings.Replace(operateInstanceURI, "{uuid}", UUID, -1)

	resp, err := c.Client.CreateRequestWithURI(http.MethodPut, realURI, requestBody)
	if err != nil {
		return nil, err
	}
	return common.GetAsyncResponse(&c.Client, resp)
}

func (c *Client) StopInstance(UUID string, stopType StopInstanceType) (*common.AsyncResponse, error) {
	requestStruct := StopInstanceRequest{}
	requestStruct.StopVMInstance.Type = stopType
	requestBody, err := json.Marshal(requestStruct)
	if err != nil {
		return nil, err
	}

	realURI := strings.Replace(operateInstanceURI, "{uuid}", UUID, -1)

	resp, err := c.Client.CreateRequestWithURI(http.MethodPut, realURI, requestBody)
	if err != nil {
		return nil, err
	}
	return common.GetAsyncResponse(&c.Client, resp)
}

func (c *Client) RestartInstance(UUID string) (*common.AsyncResponse, error) {
	requestStruct := RestartInstanceRequest{
		RebootVMInstance: map[string]string{},
	}
	requestBody, err := json.Marshal(requestStruct)
	if err != nil {
		return nil, err
	}

	realURI := strings.Replace(operateInstanceURI, "{uuid}", UUID, -1)

	resp, err := c.Client.CreateRequestWithURI(http.MethodPut, realURI, requestBody)
	if err != nil {
		return nil, err
	}
	return common.GetAsyncResponse(&c.Client, resp)
}
