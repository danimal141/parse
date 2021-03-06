package parse

import (
	"encoding/json"
	"fmt"
	"path"
	"reflect"
)

type Params map[string]interface{}

func (c *Client) CallFunction(name string, params Params, resp interface{}) error {
	return c.callFn(name, params, resp, "")
}

type callFnRequest struct {
	name   string
	params Params
	st     string
}

type fnResponse struct {
	Result interface{} `parse:"result"`
}

func (c *Client) callFn(name string, params Params, resp interface{}, sessionToken string) error {
	rv := reflect.ValueOf(resp)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("parse: expected a non-nil pointer got %v", rv.Kind())
	}

	if params == nil {
		params = Params{}
	}
	cr := &callFnRequest{
		name:   name,
		params: params,
		st:     sessionToken,
	}

	if b, err := c.doRequest(cr); err != nil {
		return err
	} else {
		r := fnResponse{}
		if err := json.Unmarshal(b, &r); err != nil {
			return err
		}
		return populateValue(resp, r.Result)
	}
}

func (c *callFnRequest) method() string {
	return "POST"
}

func (c *callFnRequest) endpoint() (string, error) {
	return path.Join("functions", c.name), nil
}

func (c *callFnRequest) body() (string, error) {
	b, err := json.Marshal(c.params)
	return string(b), err
}

func (c *callFnRequest) useMasterKey() bool {
	return false
}

func (c *callFnRequest) sessionToken() string {
	return c.st
}

func (c *callFnRequest) contentType() string {
	return "application/json"
}
