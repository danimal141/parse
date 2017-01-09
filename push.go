package parse

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Interface representing a Parse Push notification and the various
// options for sending a push notification. This API is chainable for
// conveniently building push notifications:
//
// parse.NewPushNotification().Channels("chan1", "chan2").Where(parse.NewPushQuery().EqualTo("deviceType", "ios")).Data(map[string]interface{}{"alert": "hello"}).Send()
type PushNotification interface {
	// Set the query for advanced targeting
	//
	// use parse.NewPushQuery to create a new query
	Where(q Query) PushNotification

	// Set the channels to target
	Channels(c ...string) PushNotification

	// Specify a specific time to send this push
	PushTime(t time.Time) PushNotification

	// Set the time this push notification should expire if it can't be immediately sent
	ExpirationTime(t time.Time) PushNotification

	// Set the duration after which this push notification should expire if it can't be immediately sent
	ExpirationInterval(d time.Duration) PushNotification

	// Set the payload for this push notification
	Data(d map[string]interface{}) PushNotification

	// Send the push notification
	Send() error
}

type pushRequest struct {
	client *Client

	shouldUseMasterKey bool
	channels           []string
	expirationInterval int64
	expirationTime     *Date
	pushTime           *Date
	where              map[string]interface{}
	data               map[string]interface{}
}

// Convenience function for creating a new query for use in SendPush.
func (c *Client) NewPushQuery() Query {
	q, _ := c.NewQuery(&Installation{})
	return q
}

// Create a new Push Notifaction
//
// See the Push Notification Guide for more details: https://www.parse.com/docs/push_guide#sending/REST
func (c *Client) NewPushNotification() PushNotification {
	return &pushRequest{client: c}
}

func (p *pushRequest) Where(q Query) PushNotification {
	p.where = q.(*query).where
	return p
}

func (p *pushRequest) Channels(c ...string) PushNotification {
	p.channels = c
	return p
}

func (p *pushRequest) PushTime(t time.Time) PushNotification {
	d := Date(t)
	p.pushTime = &d
	return p
}

func (p *pushRequest) ExpirationTime(t time.Time) PushNotification {
	d := Date(t)
	p.expirationTime = &d
	return p
}

func (p *pushRequest) ExpirationInterval(d time.Duration) PushNotification {
	p.expirationInterval = int64(d.Seconds())
	return p
}

func (p *pushRequest) Data(d map[string]interface{}) PushNotification {
	p.data = d
	return p
}

func (p *pushRequest) Send() error {
	b, err := p.client.doRequest(p)
	data := map[string]interface{}{}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	fmt.Printf("data: %v\n", data)
	return err
}

func (p *pushRequest) method() string {
	return "POST"
}

func (p *pushRequest) endpoint() (string, error) {
	return "push", nil
}

func (p *pushRequest) body() (string, error) {
	if p.expirationTime != nil && p.expirationInterval > 0 {
		return "", errors.New("parse: cannot use both expiration_time and expiration_interval")
	}

	payload, err := json.Marshal(&struct {
		Channels           []string               `json:"channels,omitempty"`
		ExpirationTime     *Date                  `json:"expiration_time,omitempty"`
		ExpirationInterval int64                  `json:"expiration_interval,omitempty"`
		PushTime           *Date                  `json:"push_time,omitempty"`
		Data               map[string]interface{} `json:"data,omitempty"`
		Where              map[string]interface{} `json:"where,omitempty"`
	}{
		Channels:           p.channels,
		ExpirationTime:     p.expirationTime,
		PushTime:           p.pushTime,
		ExpirationInterval: p.expirationInterval,
		Data:               p.data,
		Where:              p.where,
	})

	fmt.Printf("body: %s\n", payload)
	return string(payload), err
}

func (p *pushRequest) useMasterKey() bool {
	return p.shouldUseMasterKey
}

func (p *pushRequest) session() *session {
	return nil
}

func (p *pushRequest) contentType() string {
	return "application/json"
}
