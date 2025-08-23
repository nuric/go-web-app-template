package tests

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/glebarez/sqlite"
	"github.com/gorilla/sessions"
	"github.com/nuric/go-api-template/controllers"
	"github.com/nuric/go-api-template/email"
	"github.com/nuric/go-api-template/storage"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestLoginPage(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	config := controllers.Config{
		Database:   db,
		Session:    sessions.NewCookieStore([]byte("32-character-long-secret-key-abc")),
		Emailer:    email.LogEmailer{},
		Storer:     &storage.OsStorer{Path: t.TempDir()},
		CSRFSecret: "32-character-long-csrf-secret-key-xyz",
		Debug:      true,
	}
	handler := controllers.Setup(config)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithDebugf(t.Logf))
	defer cancel()

	err = chromedp.Run(ctx,
		chromedp.Navigate(ts.URL),
		chromedp.WaitVisible(`input[name="email"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`input[name="password"]`, chromedp.ByQuery),
	)
	require.NoError(t, err)
}
