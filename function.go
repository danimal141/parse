package parse

import (
	"encoding/json"
	"errors"
	"path"
	"reflect"
)

type Params map[string]interface{}

func (c *client) CallFunction(name string, params Params, resp interface{}) error {
	return c.callFn(name, params, resp, nil)
}

type callFnRequest struct {
	name           string
	params         Params
	currentSession *session
}

type fnResponse struct {
	Result interface{} `parse:"result"`
}

func (c *client) callFn(name string, params Params, resp interface{}, currentSession *session) error {
	rv := reflect.ValueOf(resp)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("resp must be a non-nil pointer")
	}

	if params == nil {
		params = Params{}
	}

	cr := &callFnRequest{
		name:           name,
		params:         params,
		currentSession: currentSession,
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

func (c *callFnRequest) session() *session {
	return c.currentSession
}

func (c *callFnRequest) contentType() string {
	return "application/json"
}
