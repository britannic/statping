package notifiers

import (
	"os"
	"testing"
	"time"

	"github.com/statping/statping/database"
	"github.com/statping/statping/types/notifications"
	"github.com/statping/statping/types/null"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	MOBILE_ID     string
	MOBILE_NUMBER string
)

func init() {
	MOBILE_ID = os.Getenv("MOBILE_ID")
	MOBILE_NUMBER = os.Getenv("MOBILE_NUMBER")
	Mobile.Var1 = MOBILE_ID
}

func TestMobileNotifier(t *testing.T) {
	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&notifications.Notification{})
	notifications.SetDB(db)

	Mobile.Var1 = MOBILE_ID
	Mobile.Var2 = os.Getenv("MOBILE_NUMBER")
	if MOBILE_ID == "" {
		t.Log("Mobile notifier testing skipped, missing MOBILE_ID environment variable")
		t.SkipNow()
	}

	t.Run("Load Mobile", func(t *testing.T) {
		Mobile.Var1 = MOBILE_ID
		Mobile.Var2 = MOBILE_NUMBER
		Mobile.Delay = time.Duration(100 * time.Millisecond)
		Mobile.Limits = 10
		Mobile.Enabled = null.NewNullBool(true)

		Add(Mobile)

		assert.Equal(t, "Hunter Long", Mobile.Author)
		assert.Equal(t, MOBILE_ID, Mobile.Var1)
		assert.Equal(t, MOBILE_NUMBER, Mobile.Var2)
	})

	t.Run("Mobile Notifier Tester", func(t *testing.T) {
		assert.True(t, Mobile.CanSend())
	})

	t.Run("Mobile OnFailure", func(t *testing.T) {
		err := Mobile.OnFailure(exampleService, exampleFailure)
		assert.Nil(t, err)
	})

	t.Run("Mobile OnSuccess", func(t *testing.T) {
		err := Mobile.OnSuccess(exampleService)
		assert.Nil(t, err)
	})

	t.Run("Mobile Test", func(t *testing.T) {
		t.SkipNow()
		_, err := Mobile.OnTest()
		assert.Nil(t, err)
	})

}
