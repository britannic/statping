package notifiers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/statping/statping/types/failures"
	"github.com/statping/statping/types/notifications"
	"github.com/statping/statping/types/notifier"
	"github.com/statping/statping/types/services"
	"github.com/statping/statping/utils"
)

var _ notifier.Notifier = (*telegram)(nil)

type telegram struct {
	*notifications.Notification
}

func (t *telegram) Select() *notifications.Notification {
	return t.Notification
}

var Telegram = &telegram{&notifications.Notification{
	Method:      "telegram",
	Title:       "Telegram",
	Description: "Receive notifications on your Telegram channel when a service has an issue. You must get a Telegram API token from the /botfather. Review the <a target=\"_blank\" href=\"http://techthoughts.info/how-to-create-a-telegram-bot-and-send-messages-via-api\">Telegram API Tutorial</a> to learn how to generate a new API Token.",
	Author:      "Hunter Long",
	AuthorUrl:   "https://github.com/hunterlong",
	Icon:        "fab fa-telegram-plane",
	Delay:       time.Duration(5 * time.Second),
	Limits:      60,
	Form: []notifications.NotificationForm{{
		Type:        "text",
		Title:       "Telegram API Token",
		Placeholder: "383810182:EEx829dtCeufeQYXG7CUdiQopqdmmxBPO7-s",
		SmallText:   "Enter the API Token given to you from the /botfather chat.",
		DbField:     "api_secret",
		Required:    true,
	}, {
		Type:        "text",
		Title:       "Channel or User",
		Placeholder: "@statping_channel",
		SmallText:   "Insert your Telegram Channel or User here.",
		DbField:     "var1",
		Required:    true,
	}}},
}

// Send will send a HTTP Post to the Telegram API. It accepts type: string
func (t *telegram) sendMessage(message string) (string, error) {
	apiEndpoint := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", t.ApiSecret)

	v := url.Values{}
	v.Set("chat_id", t.Var1)
	v.Set("text", message)
	rb := *strings.NewReader(v.Encode())

	contents, _, err := utils.HttpRequest(apiEndpoint, "GET", "application/x-www-form-urlencoded", nil, &rb, time.Duration(10*time.Second), true, nil)

	success, _ := telegramSuccess(contents)
	if !success {
		errorOut := telegramError(contents)
		out := fmt.Sprintf("Error code %v - %v", errorOut.ErrorCode, errorOut.Description)
		return string(contents), errors.New(out)
	}
	return string(contents), err
}

// OnFailure will trigger failing service
func (t *telegram) OnFailure(s *services.Service, f *failures.Failure) error {
	msg := fmt.Sprintf("Your service '%v' is currently offline!", s.Name)
	_, err := t.sendMessage(msg)
	return err
}

// OnSuccess will trigger successful service
func (t *telegram) OnSuccess(s *services.Service) error {
	msg := fmt.Sprintf("Your service '%v' is currently online!", s.Name)
	_, err := t.sendMessage(msg)
	return err
}

// OnTest will test the Twilio SMS messaging
func (t *telegram) OnTest() (string, error) {
	msg := fmt.Sprintf("Testing the Twilio SMS Notifier on your Statping server")
	content, err := t.sendMessage(msg)
	return content, err
}

func telegramSuccess(res []byte) (bool, telegramResponse) {
	var obj telegramResponse
	json.Unmarshal(res, &obj)
	if obj.Ok {
		return true, obj
	}
	return false, obj
}

func telegramError(res []byte) telegramErrorObj {
	var obj telegramErrorObj
	json.Unmarshal(res, &obj)
	return obj
}

type telegramErrorObj struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

type telegramResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID        int    `json:"id"`
			IsBot     bool   `json:"is_bot"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID        int    `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Username  string `json:"username"`
			Type      string `json:"type"`
		} `json:"chat"`
		Date int    `json:"date"`
		Text string `json:"text"`
	} `json:"result"`
}
