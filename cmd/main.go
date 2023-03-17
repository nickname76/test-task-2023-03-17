package main

import (
	"encoding/json"
	"fmt"

	cdekcalc "github.com/nickname76/test-task-2023-03-17"
)

func main() {
	testAccount := "EMscd6r9JnFiQ3bLoyjJY6eM78JrJceI"
	testSecurePassword := "PjLZkKBHEiLK3YsjtNrt3TGNG0ahs3kG"
	testMode := true
	token, _, err := cdekcalc.GetToken(testAccount, testSecurePassword, testMode, "")
	if err != nil {
		panic(err)
	}

	client := cdekcalc.NewClient(token, true, "")

	addrFrom := "Россия, г. Москва, Cлавянский бульвар д.1"
	addrTo := "Россия, Воронежская обл., г. Воронеж, ул. Ленина д.43"
	size := cdekcalc.Size{
		Weight: 100,
		Length: 10,
		Width:  10,
		Height: 10,
	}
	prices, errors, err := client.Calculate(addrFrom, addrTo, size)
	if err != nil {
		panic(err)
	}

	if errors != nil {
		b, err := json.MarshalIndent(errors, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))

		return
	}

	b, err := json.MarshalIndent(prices, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
