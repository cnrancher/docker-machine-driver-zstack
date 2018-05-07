package instance

import (
	"github.com/orangedeng/go-zstack/common"
	"encoding/json"
	"net/http"
)

const (
	createOfferingURI = "/v1/instance-offerings"
)

type Offering struct {
	common.Client
}

func (c *Offering)CreateOffering(req CreateOfferingRequest) (*common.AsyncResponse, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.CreateRequestWithURI(http.MethodPost, createOfferingURI, requestBody)
	if err != nil {
		return nil, err
	}

	return common.GetAsyncResponse(&c.Client, resp)
}
