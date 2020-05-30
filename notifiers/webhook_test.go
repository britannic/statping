package notifiers

import (
	"testing"

	"github.com/statping/statping/database"
	"github.com/statping/statping/types/notifications"
	"github.com/statping/statping/types/null"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	webhookTestUrl = "https://statping.com"
	webhookMessage = `{"id": {{.Service.Id}},"name": "{{.Service.Name}}","online": {{.Service.Online}},"issue": "{{.Failure.Issue}}"}`
	apiKey         = "application/json"
	fullMsg        string
)

func TestWebhookNotifier(t *testing.T) {
	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)

	t.Run("Load webhooker", func(t *testing.T) {
		Webhook.Host = webhookTestUrl
		Webhook.Var1 = "POST"
		Webhook.Var2 = webhookMessage
		Webhook.ApiKey = "application/json"
		Webhook.Enabled = null.NewNullBool(true)

		Add(Webhook)

		assert.Equal(t, "Hunter Long", Webhook.Author)
		assert.Equal(t, webhookTestUrl, Webhook.Host)
		assert.Equal(t, apiKey, Webhook.ApiKey)
	})

	t.Run("webhooker Notifier Tester", func(t *testing.T) {
		assert.True(t, Webhook.CanSend())
	})

	t.Run("webhooker OnFailure", func(t *testing.T) {
		err := Webhook.OnFailure(exampleService, exampleFailure)
		assert.Nil(t, err)
	})

	t.Run("webhooker OnSuccess", func(t *testing.T) {
		err := Webhook.OnSuccess(exampleService)
		assert.Nil(t, err)
	})

	t.Run("webhooker Send", func(t *testing.T) {
		err := Webhook.Send(fullMsg)
		assert.Nil(t, err)
	})

}
