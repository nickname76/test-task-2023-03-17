package cdekcalc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	AuthEndpointURLProduction = "https://api.cdek.ru/v2/oauth/token"
	AuthEndpointURLTesting    = "https://api.edu.cdek.ru/v2/oauth/token"
)

// Ответ /oauth/token
//
// https://api-docs.cdek.ru/29923918.html
type respOAuthToken struct {
	// jwt-токен
	AccessToken string `json:"access_token,omitempty"`
	// срок действия токена (по умолчанию 3600 секунд)
	ExpiresIn int `json:"expires_in,omitempty"`

	// поля token_type, scope, jti пропущены

	// Код ошибки
	Error string `json:"error,omitempty"`
	// Описание ошибки
	ErrorDescription string `json:"error_description,omitempty"`
}

// Получает OAuth токен (для Client) из Account и Secure password.
//
// Возвращает сам token и expiresIn - время жизни токена в секундах.
//
// Укажите testMode = true, если хотите использовать тестовый сервер,
// или укажите свой в customAuthEndpointURL, введя полный адрес метода oauth/token
//
// https://api-docs.cdek.ru/29923918.html
func GetToken(account, securePassword string, testMode bool, customAuthEndpointURL string) (token string, expiresIn int, err error) {
	endpointURL := ""

	if customAuthEndpointURL != "" {
		endpointURL = customAuthEndpointURL
	} else {
		if testMode {
			endpointURL = AuthEndpointURLTesting
		} else {
			endpointURL = AuthEndpointURLProduction
		}
	}

	params := url.Values{}
	params.Set("grant_type", "client_credentials")
	params.Set("client_id", account)
	params.Set("client_secret", securePassword)

	resp, err := http.PostForm(endpointURL, params)
	if err != nil {
		return "", 0, fmt.Errorf("GetAccessToken: %w", err)
	}
	defer resp.Body.Close()

	respData := new(respOAuthToken)
	err = json.NewDecoder(resp.Body).Decode(respData)
	if err != nil {
		return "", 0, fmt.Errorf("GetAccessToken: %w", err)
	}

	if respData.Error != "" {
		return "", 0, fmt.Errorf("GetAccessToken: API error '%w' (%v)", errors.New(respData.Error), respData.ErrorDescription)

	}

	return respData.AccessToken, respData.ExpiresIn, nil
}
