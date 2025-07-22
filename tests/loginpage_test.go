package tests

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/glebarez/sqlite"
	"github.com/gorilla/sessions"
	"github.com/nuric/go-api-template/routes"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestLoginPage(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	ss := sessions.NewCookieStore([]byte("32-character-long-secret-key-abc"))
	routes := routes.SetupRoutes(db, ss)
	ts := httptest.NewServer(routes)
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
