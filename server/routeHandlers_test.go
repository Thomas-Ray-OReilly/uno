package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func createMockServerAndRequest() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	setupRoutes(e)
	req := httptest.NewRequest(http.MethodPost, "/newgame", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return c, rec
}

//TestRespondIfValid this wont bring the code coverage up but its worth it to test this if it breaks on this level.
func TestRespondIfValid(t *testing.T) {
	// Setup
	e := echo.New()
	setupRoutes(e)
	req := httptest.NewRequest(http.MethodGet, "/respondIfValid", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, respondIfValid(c, true, "", "")) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestNewGame(t *testing.T) {
	// Setup
	c, rec := createMockServerAndRequest()

	// Assertions
	if assert.NoError(t, newGame(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		// store response in map and make sure valid field is true
		var recData map[string]interface{}
		json.Unmarshal([]byte(rec.Body.String()), &recData)
		assert.Equal(t, true, recData["valid"])
	}
}

func TestLogin(t *testing.T) {
	// Setup
	e := echo.New()
	setupRoutes(e)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/login/0/tester_name/false", nil))

	var res Response
	json.Unmarshal([]byte(rec.Body.String()), &res)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, res.ValidGame, true)
	// JWT present
	assert.NotEqual(t, res.Payload["JWT"], nil)
}

func TestDraw(t *testing.T) {
	// Setup
	e := echo.New()
	setupRoutes(e)

	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, httptest.NewRequest(http.MethodPost, "/login/0/tester_name/false", nil))

	var loginRes Response
	json.Unmarshal([]byte(loginRec.Body.String()), &loginRes)
	assert.NotEqual(t, loginRes.Payload["JWT"], nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/draw", nil)
	req.Header.Set("Authorization", "Bearer "+loginRes.Payload["JWT"].(string))
	e.ServeHTTP(rec, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdate(t *testing.T) {
	// Setup
	e := echo.New()
	setupRoutes(e)

	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, httptest.NewRequest(http.MethodPost, "/login/0/tester_name/false", nil))

	var loginRes Response
	json.Unmarshal([]byte(loginRec.Body.String()), &loginRes)
	assert.NotEqual(t, loginRes.Payload["JWT"], nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/update", nil)
	req.Header.Set("Authorization", "Bearer "+loginRes.Payload["JWT"].(string))
	e.ServeHTTP(rec, req)

	var res Response
	json.Unmarshal([]byte(rec.Body.String()), &res)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPlay(t *testing.T) {
	// Setup - before you can play you must login and start a game
	e := echo.New()
	setupRoutes(e)
	// login
	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, httptest.NewRequest(http.MethodPost, "/login/0/tester_name/false", nil))
	var loginRes Response
	json.Unmarshal([]byte(loginRec.Body.String()), &loginRes)
	assert.NotEqual(t, loginRes.Payload["JWT"], nil)

	playRec := httptest.NewRecorder()
	playReq := httptest.NewRequest(http.MethodPost, "/play/1/blue", nil)
	playReq.Header.Set("Authorization", "Bearer "+loginRes.Payload["JWT"].(string))
	e.ServeHTTP(playRec, playReq)

	assert.Equal(t, http.StatusOK, playRec.Code)
}

func TestStartGame(t *testing.T) {
	// Setup - you must login to start a game
	e := echo.New()
	setupRoutes(e)
	// login
	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, httptest.NewRequest(http.MethodPost, "/login/0/tester_name/true", nil))
	var loginRes Response
	json.Unmarshal([]byte(loginRec.Body.String()), &loginRes)
	assert.NotEqual(t, loginRes.Payload["JWT"], nil)

	// TODO currently broken - players array seem to be empty after login
	// startRec := httptest.NewRecorder()
	// startReq := httptest.NewRequest(http.MethodPost, "/startgame", nil)
	// startReq.Header.Set("Authorization", "Bearer "+loginRes.Payload["JWT"].(string))
	// e.ServeHTTP(startRec, startReq)
	// assert.Equal(t, http.StatusOK, startRec.Code)
}
