package routes_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/nuric/go-api-template/models"
	"github.com/nuric/go-api-template/routes"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func makeRequest(t *testing.T, method string, endpoint string, body any, resp any) int {
	t.Helper()
	// ---------------------------
	var bodyReader io.Reader
	if body != nil {
		var bodyAsString string
		if bodyString, ok := body.(string); ok {
			bodyAsString = bodyString
		} else {
			jsonBody, err := json.Marshal(body)
			require.NoError(t, err)
			bodyAsString = string(jsonBody)
		}
		bodyReader = bytes.NewReader([]byte(bodyAsString))
	}
	req, err := http.NewRequest(method, endpoint, bodyReader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	// ---------------------------
	recorder := httptest.NewRecorder()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}))
	handler := routes.SetupRoutes(db)
	handler.ServeHTTP(recorder, req)
	// ---------------------------
	if resp != nil {
		err = json.Unmarshal(recorder.Body.Bytes(), resp)
		require.NoError(t, err)
	}
	t.Log(recorder.Body.String())
	// ---------------------------
	return recorder.Code
}

func Test_Greetings(t *testing.T) {
	// ---------------------------
	reqBody := routes.GreetingRequest{
		LastName: "Wizard",
	}
	var respBody routes.GreetingResponse
	resp := makeRequest(t, http.MethodPost, "/greetings/Gandalf", reqBody, &respBody)
	require.Equal(t, http.StatusOK, resp)
	require.Equal(t, "Hello, Gandalf Wizard!", respBody.Greeting)
	// ---------------------------
	// Empty first name should be at least 2 characters
	var errResp map[string]string
	resp = makeRequest(t, http.MethodPost, "/greetings/g", reqBody, &errResp)
	require.Equal(t, http.StatusBadRequest, resp)
	require.Equal(t, "first name must be at least 2 characters", errResp["error"])
	// ---------------------------
	// Empty last name should return 400
	reqBody.LastName = ""
	resp = makeRequest(t, http.MethodPost, "/greetings/Gandalf", reqBody, nil)
	require.Equal(t, http.StatusBadRequest, resp)
	// ---------------------------
}
