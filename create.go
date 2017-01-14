package parse

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type createRequest struct {
	v                  interface{}
	shouldUseMasterKey bool
	st                 string

	isUser   bool
	username string
	password string
}

// Save a new instance of the type pointed to by v to the Parse database. If
// useMasteKey=true, the Master Key will be used for the creation request. On a
// successful request, the CreatedAt field will be set on v.
//
// Note: v should be a pointer to a struct whose name represents a Parse class,
// or that implements the ClassName method
func (c *Client) Create(v interface{}, useMasterKey bool) error {
	return c.create(v, useMasterKey, "")
}

func (c *Client) Signup(username string, password string, user interface{}) error {
	cr := &createRequest{
		v:                  user,
		shouldUseMasterKey: false,
		st:                 "",
		isUser:             true,
		username:           username,
		password:           password,
	}
	if b, err := c.doRequest(cr); err != nil {
		return err
	} else {
		return handleResponse(b, user)
	}
}

func (c *Client) create(v interface{}, useMasterKey bool, sessionToken string) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("parse: expected a non-nil pointer got %v", rv.Kind())
	}

	cr := &createRequest{
		v:                  v,
		shouldUseMasterKey: useMasterKey,
		st:                 sessionToken,
	}
	if b, err := c.doRequest(cr); err != nil {
		return err
	} else {
		return handleResponse(b, v)
	}
}

func (c *createRequest) method() string {
	return "POST"
}

func (c *createRequest) endpoint() (string, error) {
	return getEndpointBase(c.v), nil
}

func (c *createRequest) body() (string, error) {
	payload := map[string]interface{}{}

	if c.isUser {
		payload["username"] = c.username
		payload["password"] = c.password
	}

	rv := reflect.ValueOf(c.v)
	rvi := reflect.Indirect(rv)
	rt := rvi.Type()
	fields := getFields(rt)

	for _, f := range fields {
		var name string
		var fv reflect.Value

		if n, o := parseTag(f.Tag.Get("parse")); n == "-" || n == "objectId" || f.Name == "Id" || f.Type == reflect.TypeOf(Base{}) {
			continue
		} else if fv = rvi.FieldByName(f.Name); !fv.IsValid() || o == "omitempty" && isEmptyValue(fv) {
			continue
		} else {
			name = n
		}

		var fname string
		if name != "" {
			fname = name
		} else {
			fname = firstToLower(f.Name)
		}
		if canBeNil(fv) && fv.IsNil() {
			payload[fname] = nil
		} else {
			payload[fname] = encodeForRequest(fv.Interface())
		}
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *createRequest) useMasterKey() bool {
	return c.shouldUseMasterKey
}

func (c *createRequest) sessionToken() string {
	return c.st
}

func (c *createRequest) contentType() string {
	return "application/json"
}
