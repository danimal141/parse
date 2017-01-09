package parse

import (
	"fmt"
	"path"
	"reflect"
)

// Delete the instance of the type represented by v from the Parse database. If
// useMasteKey=true, the Master Key will be used for the deletion request.
func (c *Client) Delete(v interface{}, useMasterKey bool) error {
	return c._delete(v, useMasterKey, nil)
}

func (c *Client) _delete(v interface{}, useMasterKey bool, currentSession *session) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("parse: expected a non-nil pointer got %v", rv.Kind())
	}

	_, err := c.doRequest(&deleteRequest{
		inst:               v,
		shouldUseMasterKey: useMasterKey,
		currentSession:     currentSession,
	})
	return err
}

type deleteRequest struct {
	inst               interface{}
	shouldUseMasterKey bool
	currentSession     *session
}

func (d *deleteRequest) method() string {
	return "DELETE"
}

func (d *deleteRequest) endpoint() (string, error) {
	var id string
	rv := reflect.ValueOf(d.inst)
	rvi := reflect.Indirect(rv)
	if f := rvi.FieldByName("Id"); f.IsValid() {
		if s, ok := f.Interface().(string); ok {
			id = s
		} else {
			return "", fmt.Errorf("parse: Id field should be a string, received type %s", f.Type())
		}
	} else {
		return "", fmt.Errorf("parse: can not delete value - type has no Id field")
	}

	return path.Join(getEndpointBase(d.inst), id), nil
}

func (d *deleteRequest) body() (string, error) {
	return "", nil
}

func (d *deleteRequest) useMasterKey() bool {
	return d.shouldUseMasterKey
}

func (d *deleteRequest) session() *session {
	return d.currentSession
}

func (d *deleteRequest) contentType() string {
	return "application/x-www-form-urlencoded"
}
