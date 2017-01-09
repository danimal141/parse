package parse

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestDeleteRequiresPointer(t *testing.T) {
	u := User{}
	rv := reflect.TypeOf(u)
	expected := fmt.Sprintf("parse: expected a non-nil pointer got %v", rv.Kind())
	if err := testClient.Delete(u, true); err == nil {
		t.Error("Delete should return an error when argument is not a pointer")
	} else if err.Error() != expected {
		t.Errorf("Unexpected error message. Got [%s] expected [%s]\n", err, expected)
	}

	if err := testClient.Delete(u, false); err == nil {
		t.Error("Delete should return an error when argument is not a pointer")
	} else if err.Error() != expected {
		t.Errorf("Unexpected error message. Got [%s] expected [%s]\n", err, expected)
	}
}

func TestEndpointDelete(t *testing.T) {
	testCases := []struct {
		inst     interface{}
		id       string
		expected string
	}{
		{&User{Base: Base{Id: "UserId1"}}, "UserId1", "users/UserId1"},
		{&CustomClass{Base{Id: "Custom1"}}, "Custom1", "classes/CustomClass/Custom1"},
		{&CustomClassCustomName{Base{Id: "CC2"}}, "CC2", "classes/customName/CC2"},
		{&CustomClassCustomEndpoint{Base{Id: "Cc3"}}, "Cc3", "custom/class/endpoint/Cc3"},
	}

	for _, tc := range testCases {
		d := &deleteRequest{inst: tc.inst}
		actual, err := d.endpoint()
		if err != nil {
			t.Errorf("Unexpected error creating query: %v\n", err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("Wrong endpoint generated. Expected [%s] got [%s]\n", tc.expected, actual)
		}
	}
}

func TestDelete(t *testing.T) {
	shouldHaveMasterKey := false
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get(AppIdHeader); h != "app_id" {
			t.Errorf("request did not have App ID header set!")
		}

		if h := r.Header.Get(SessionTokenHeader); h != "" {
			t.Errorf("request had Session Token header set!")
		}

		if shouldHaveMasterKey {
			if h := r.Header.Get(RestKeyHeader); h != "" {
				t.Errorf("request had Rest Key header set!")
			}

			if h := r.Header.Get(MasterKeyHeader); h != "master_key" {
				t.Errorf("request did not have Master Key header set!")
			}
		} else {
			if h := r.Header.Get(RestKeyHeader); h != "rest_key" {
				t.Errorf("request did not have Rest Key header set!")
			}

			if h := r.Header.Get(MasterKeyHeader); h != "" {
				t.Errorf("request had Master Key header set!")
			}
		}

		fmt.Fprintf(w, "")
	})
	defer teardownTestServer()

	u := User{Base: Base{Id: "abc"}}
	testClient.Delete(&u, false)
}
