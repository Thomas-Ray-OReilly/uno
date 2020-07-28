package main

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jak103/uno/model"
	"github.com/labstack/echo/v4"
)

var sim bool = true

type Response struct {
	ValidGame bool                   `json:"valid"` // Valid game id/game id is in JWT
	Payload   map[string]interface{} `json:"payload"`
}

func setupRoutes(e *echo.Echo) {
	e.GET("/newgame/:username", newGame)
	e.GET("/update", update)
	e.POST("/startgame", startGame)
	e.POST("/login/:game/:username", login)
	e.POST("/play/:number/:color", play)
	e.POST("/draw", draw)
}

func newGame(c echo.Context) error {
	gameid, gameErr := createNewGame()

	if gameErr != nil {
		return gameErr
	}

	// TODO: validate username
	encodedJWT, err := newJWT(c.Param("username"), uuid.New(), gameid, true, []byte(signKey))

	payload := newPayload(c.Param("username"), gameid)

	if err == nil {
		payload = MakeJWTPayload(payload, encodedJWT)
	} else {
		return c.JSONPretty(http.StatusNonAuthoritativeInfo, &Response{false, nil}, " ")
	}

	return c.JSONPretty(http.StatusOK, &Response{true, payload}, "  ")
}

func login(c echo.Context) error {
	fmt.Println(c.Param("game"))
	fmt.Println(c.Param("username"))
	err := joinGame(c.Param("game"), c.Param("username"))
	return respondWithJWTIfValid(c, err)
}

func startGame(c echo.Context) error {
	dealCards()
	return update(c)
}

func update(c echo.Context) error {

	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
	if authHeader == "" {
		return c.JSONPretty(http.StatusUnauthorized, &Response{false, nil}, " ")
	}
	claims, validUser := getValidClaims(authHeader)

	if !validUser {
		return c.JSONPretty(http.StatusUnauthorized, &Response{false, nil}, " ")
	}

	valid := updateGame(claims["gameid"].(string))
	return respondIfValid(c, valid && validUser, claims["name"].(string), claims["gameid"].(string))
}

func play(c echo.Context) error {

	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
	if authHeader == "" {
		return c.JSONPretty(http.StatusUnauthorized, &Response{false, nil}, " ")
	}
	claims, validUser := getValidClaims(authHeader)

	if !validUser {
		return c.JSONPretty(http.StatusUnauthorized, &Response{false, nil}, " ")
	}

	// TODO Cards have a value, which can include skip, reverse, etc
	card := model.Card{c.Param("number"), c.Param("color")}
	valid := playCard(claims["gameid"].(string), claims["name"].(string), card)
	return respondIfValid(c, valid, claims["name"].(string), claims["gameid"].(string))
}

func draw(c echo.Context) error {

	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
	if authHeader == "" {
		return c.JSONPretty(http.StatusUnauthorized, &Response{false, nil}, " ")
	}
	claims, validUser := getValidClaims(authHeader)

	if !validUser {
		return c.JSONPretty(http.StatusUnauthorized, &Response{false, nil}, " ")
	}

	valid := drawCard(claims["gameid"].(string), claims["name"].(string))
	return respondIfValid(c, valid, claims["name"].(string), claims["gameid"].(string))
}

func respondIfValid(c echo.Context, valid bool, username string, gameId string) error {
	var response *Response
	if valid {
		response = &Response{true, newPayload(username, gameId)}
	} else {
		response = &Response{false, nil}
	}
	return c.JSONPretty(http.StatusOK, response, "  ")
}

func respondWithJWTIfValid(c echo.Context, optInputError error) error {
	// TODO: validate username and game id

	// Check if they have a JWT before just overriding it!
	// If they do, we need to make a JWT based off of their current one, but add/change the gameid.
	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
	var username string = c.Param("username")
	userId := uuid.New()
	var host bool = false
	if authHeader != "" {
		claims, validUser := getValidClaims(authHeader)
		if validUser {
			username = claims["name"].(string)
			userId = claims["userid"].(uuid.UUID)
			host = claims["isHost"].(bool)
		}
	}

	encodedJWT, err := newJWT(username, userId, c.Param("game"), host, []byte(signKey))

	payload := newPayload(username, c.Param("game"))

	if err == nil {
		payload = MakeJWTPayload(payload, encodedJWT)
	} else {
		return c.JSONPretty(http.StatusNonAuthoritativeInfo, &Response{false, nil}, " ")
	}

	var response *Response

	status := http.StatusOK

	if optInputError == nil {
		response = &Response{true, payload}
	} else {
		// forward the error if any
		payload := make(map[string]interface{})
		payload["error"] = optInputError.Error()
		response = &Response{false, payload}
		status = http.StatusBadRequest
	}

	return c.JSONPretty(status, response, "  ")
}
