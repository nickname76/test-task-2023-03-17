package cdekcalc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	EndpointURLProduction = "https://api.cdek.ru/v2/calculator/tarifflist"
	EndpointURLTesting    = "https://api.edu.cdek.ru/v2/calculator/tarifflist"
)

type Client struct {
	Token      string
	EnpointURL string
}

// Создаёт новый клиент для получения расчётов по доступным тарифам.
//
// Принимает:
//   - token - токен для СДЕК API, можно получить в GetToken
//   - testMode - true, если должен использовать тестовый сервер
//   - customEndpointURL - свой URL метода /calculator (если указан, значение testMode не будет изменять URL метода)
func NewClient(token string, testMode bool, customEndpointURL string) *Client {
	api := new(Client)

	api.Token = token

	if customEndpointURL != "" {
		api.EnpointURL = customEndpointURL
	} else {
		if testMode {
			api.EnpointURL = EndpointURLTesting
		} else {
			api.EnpointURL = EndpointURLProduction
		}
	}

	return api
}

// Информация о габаритах груза
type Size struct {
	// Общий вес (в граммах)
	Weight int `json:"weight"`
	// Не обязательно. Габариты упаковки. Длина (в сантиметрах)
	Length int `json:"length,omitempty"`
	// Не обязательно. Габариты упаковки. Ширина (в сантиметрах)
	Width int `json:"width,omitempty"`
	// Не обязательно. Габариты упаковки. Высота (в сантиметрах)
	Height int `json:"height,omitempty"`
}

// Ответ на расчет по доступным тарифам
type PriceSending struct {
	// Код тарифа (подробнее см. [приложение 2])
	TariffCode int `json:"tariff_code"`
	// Название тарифа на языке вывода
	TariffName string `json:"tariff_name"`
	// Описание тарифа на языке вывода
	TariffDescription string `json:"tariff_description"`
	// Режим тарифа (подробнее см. приложение 3)
	DeliveryMode int `json:"delivery_mode"`
	// Стоимость доставки
	DeliverySum float64 `json:"delivery_sum"`
	// Минимальное время доставки (в рабочих днях)
	PeriodMin int `json:"period_min"`
	// Максимальное время доставки (в рабочих днях)
	PeriodMax int `json:"period_max"`
	// Может быть не указано. Минимальное время доставки (в календарных днях)
	CalendarMin int `json:"calendar_min,omitempty"`
	// Может быть не указано.  Максимальное время доставки (в календарных днях)
	CalendarMax int `json:"calendar_max,omitempty"`
}

// Ошибка в запросе на расчет по доступным тарифам
type Error struct {
	// Код ошибки
	Code string `json:"code"`
	// Описание ошибки
	Message string `json:"message"`
}

// Запрос к /calculator
//
// Неиспользуемые в библиотеке поля пропущены.
//
// https://api-docs.cdek.ru/63345519.html
type reqCalculator struct {
	// Адрес отправления
	FromLocation struct {
		// Полная строка адреса
		Address string `json:"address"`
	} `json:"from_location"`
	// Адрес получения
	ToLocation struct {
		// Полная строка адреса
		Address string `json:"address"`
	} `json:"to_location"`
	// Список информации по местам (упаковкам)
	Packages []Size `json:"packages"`
}

// Ответ /calculator
//
// https://api-docs.cdek.ru/63345519.html
type respCalculator struct {
	// Доступные тарифы
	TariffCodes []PriceSending `json:"tariff_codes,omitempty"`
	// Список ошибок
	Errors []Error `json:"errors,omitempty"`

	// Возвращается, если была ошибка во время
	// запроса, не относящаяся к методу /calculator, например в
	// случае неправильного токена
	Requests []struct {
		Errors []Error `json:"errors"`
	} `json:"requests,omitempty"`
}

// Калькулятор. Расчет по доступным тарифам
//
// Метод используется клиентами для расчета стоимости и сроков доставки по всем доступным тарифам.
//
// Принимает:
//   - addFrom - полный адрес отправления
//   - addrTo - полный адрес получение
//   - size - габариты посылки
//
// https://api-docs.cdek.ru/63345519.html
func (client *Client) Calculate(addrFrom string, addrTo string, size Size) ([]PriceSending, []Error, error) {
	reqData := &reqCalculator{
		FromLocation: struct {
			Address string "json:\"address\""
		}{
			Address: addrFrom,
		},
		ToLocation: struct {
			Address string "json:\"address\""
		}{
			Address: addrTo,
		},
		Packages: []Size{size},
	}

	buf := bytes.NewBuffer(nil)

	err := json.NewEncoder(buf).Encode(reqData)
	if err != nil {
		return nil, nil, fmt.Errorf("Calculate: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, client.EnpointURL, buf)
	if err != nil {
		return nil, nil, fmt.Errorf("Calculate: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+client.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("Calculate: %w", err)
	}
	defer resp.Body.Close()

	respData := new(respCalculator)

	err = json.NewDecoder(resp.Body).Decode(respData)
	if err != nil {
		return nil, nil, fmt.Errorf("Calculate: %w", err)
	}

	if len(respData.Errors) != 0 {
		return nil, respData.Errors, nil
	}
	if len(respData.Requests) != 0 {
		e := []Error{}
		for _, r := range respData.Requests {
			e = append(e, r.Errors...)
		}

		return nil, e, nil
	}

	return respData.TariffCodes, nil, nil
}
