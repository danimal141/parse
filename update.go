package parse

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"reflect"
)

type updateRequestype int

const (
	opSet updateRequestype = iota
	opIncr
	opDelete
	opAdd
	opAddUnique
	opRemove
	opAddRelation
	opRemoveRelation
)

func (u updateRequestype) String() string {
	switch u {
	case opSet:
		return "Set"
	case opIncr:
		return "Increment"
	case opDelete:
		return "Delete"
	case opAdd:
		return "Add"
	case opAddUnique:
		return "AddUnique"
	case opRemove:
		return "Remove"
	case opAddRelation:
		return "AddRelation"
	case opRemoveRelation:
		return "RemoveRelation"
	}
	return "Unknown"
}

func (u updateRequestype) argKey() string {
	switch u {
	case opIncr:
		return "amount"
	case opAdd, opAddUnique, opRemove, opAddRelation, opRemoveRelation:
		return "objects"
	}
	return "unknown"
}

type updateOp struct {
	UpdateType updateRequestype
	Value      interface{}
}

func (u updateOp) MarshalJSON() ([]byte, error) {
	switch u.UpdateType {
	case opSet:
		return json.Marshal(u.Value)
	case opDelete:
		return json.Marshal(map[string]interface{}{
			"__op": u.UpdateType.String(),
		})
	default:
		return json.Marshal(map[string]interface{}{
			"__op":                u.UpdateType.String(),
			u.UpdateType.argKey(): u.Value,
		})
	}
}

type Update interface {

	//Set the field specified by f to the value of v
	Set(f string, v interface{})

	// Increment the field specified by f by the amount specified by v.
	// v should be a numeric type
	Increment(f string, v interface{})

	// Delete the field specified by f from the instance being updated
	Delete(f string)

	// Append the values provided to the Array field specified by f. This operation
	// is atomic
	Add(f string, vs ...interface{})

	// Add any values provided that were not alread present to the Array field
	// specified by f. This operation is atomic
	AddUnique(f string, vs ...interface{})

	// Remove the provided values from the array field specified by f
	Remove(f string, vs ...interface{})

	// Update the ACL on the given object
	SetACL(a ACL)

	// Use the Master Key for this update request
	UseMasterKey()

	// Set the session token for the given request.
	SetSessionToken(st string)

	// Execute this update. This method also updates the proper fields
	// on the provided value with their repective new values
	Execute() error

	request
}

type updateRequest struct {
	client *Client

	inst               interface{}
	values             map[string]updateOp
	st                 string
	shouldUseMasterKey bool
}

// Create a new update request for the Parse object represented by v.
//
// Note: v should be a pointer to a struct whose name represents a Parse class,
// or that implements the ClassName method
func (c *Client) NewUpdate(v interface{}) (Update, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, fmt.Errorf("parse: expected a non-nil pointer got %v", rv.Kind())
	}

	return &updateRequest{
		client: c,
		inst:   v,
		values: map[string]updateOp{},
	}, nil
}

func (u *updateRequest) Set(f string, v interface{}) {
	u.values[f] = updateOp{UpdateType: opSet, Value: encodeForRequest(v)}
}

func (u *updateRequest) Increment(f string, v interface{}) {
	u.values[f] = updateOp{UpdateType: opIncr, Value: v}
}

func (u *updateRequest) Delete(f string) {
	u.values[f] = updateOp{UpdateType: opDelete}
}

func (u *updateRequest) Add(f string, vs ...interface{}) {
	u.values[f] = updateOp{UpdateType: opAdd, Value: vs}
}

func (u *updateRequest) AddUnique(f string, vs ...interface{}) {
	u.values[f] = updateOp{UpdateType: opAddUnique, Value: vs}
}

func (u *updateRequest) Remove(f string, vs ...interface{}) {
	u.values[f] = updateOp{UpdateType: opRemove, Value: vs}
}

func (u *updateRequest) SetACL(a ACL) {
	u.values["ACL"] = updateOp{UpdateType: opSet, Value: a}
}

func (u *updateRequest) Execute() (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("parse: error executing update: %v", r)
			}
		}
	}()

	rv := reflect.ValueOf(u.inst)
	rvi := reflect.Indirect(rv)
	fieldMap := getFieldNameMap(rv)

	for k, v := range u.values {
		var fname string
		if fn, ok := fieldMap[k]; ok {
			fname = fn
		} else {
			fname = k
		}
		fname = firstToUpper(fname)
		dv := reflect.ValueOf(v.Value)
		dvi := reflect.Indirect(dv)

		if fv := rvi.FieldByName(fname); fv.IsValid() {
			fvi := reflect.Indirect(fv)

			switch v.UpdateType {
			case opSet:
				if fv.Kind() == reflect.Ptr && fv.IsNil() && v.Value != nil {
					fv.Set(reflect.New(fv.Type().Elem()))
				}

				var tmp reflect.Value
				if fv.Kind() == reflect.Ptr {
					if v.Value == nil {
						tmp = fv.Addr()
					} else {
						tmp = fv
					}
				} else {
					tmp = fv.Addr()
				}
				if err := populateValue(tmp.Interface(), v.Value); err != nil {
					return err
				}
			case opIncr:
				switch fvi.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if dvi.Type().ConvertibleTo(fvi.Type()) {
						current := fvi.Int()
						amount := dvi.Convert(fvi.Type()).Int()
						current += amount
						fvi.Set(reflect.ValueOf(current).Convert(fvi.Type()))
					}
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if dvi.Type().ConvertibleTo(fvi.Type()) {
						current := fvi.Uint()
						amount := dvi.Convert(fvi.Type()).Uint()
						current += amount
						fvi.Set(reflect.ValueOf(current).Convert(fvi.Type()))
					}
				case reflect.Float32, reflect.Float64:
					if dvi.Type().ConvertibleTo(fvi.Type()) {
						current := fvi.Float()
						amount := dvi.Convert(fvi.Type()).Float()
						current += amount
						fvi.Set(reflect.ValueOf(current).Convert(fvi.Type()))
					}
				}
			case opDelete:
				fv.Set(reflect.Zero(fv.Type()))
			}
		}
	}
	if b, err := u.client.doRequest(u); err != nil {
		return err
	} else {
		return handleResponse(b, u.inst)
	}
}

func (u *updateRequest) UseMasterKey() {
	u.shouldUseMasterKey = true
}

func (u *updateRequest) SetSessionToken(st string) {
	u.st = st
}

func (u *updateRequest) method() string {
	return "PUT"
}

func (u *updateRequest) endpoint() (string, error) {
	p := getEndpointBase(u.inst)
	rv := reflect.ValueOf(u.inst)
	rvi := reflect.Indirect(rv)

	if f := rvi.FieldByName("Id"); f.IsValid() {
		if s, ok := f.Interface().(string); ok {
			p = path.Join(p, s)
		} else {
			return "", fmt.Errorf("parse: Id field should be a string, got type %s", f.Type())
		}
	} else {
		return "", fmt.Errorf("parse: can not update value - type has no Id field")
	}
	return p, nil
}

func (u *updateRequest) body() (string, error) {
	b, err := json.Marshal(u.values)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (u *updateRequest) useMasterKey() bool {
	return u.shouldUseMasterKey
}

func (u *updateRequest) sessionToken() string {
	return u.st
}

func (u *updateRequest) contentType() string {
	return "application/json"
}

func (c *Client) LinkFacebookAccount(u *User, a *FacebookAuthData) error {
	if u.Id == "" {
		return errors.New("parse: user Id field must not be empty")
	}
	up, _ := c.NewUpdate(u)
	up.Set("authData", AuthData{Facebook: a})
	up.UseMasterKey()
	return up.Execute()
}
