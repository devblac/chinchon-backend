//go:build tinygo
// +build tinygo

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/marianogappa/chinchon-backend/chinchon"
	"github.com/marianogappa/chinchon-backend/examplebot/newbot"
)

func main() {
	js.Global().Set("chinchonNew", js.FuncOf(chinchonNew))
	js.Global().Set("chinchonRunAction", js.FuncOf(chinchonRunAction))
	js.Global().Set("chinchonBotRunAction", js.FuncOf(chinchonBotRunAction))
	select {}
}

var (
	state *chinchon.GameState
	bot   chinchon.Bot
)

type rules struct {
	MaxPoints     int  `json:"maxPoints"`
	IsFlorEnabled bool `json:"isFlorEnabled"`
}

func chinchonNew(this js.Value, p []js.Value) interface{} {
	jsonBytes := make([]byte, p[0].Length())
	js.CopyBytesToGo(jsonBytes, p[0])
	var r rules
	// ignore rules if unmarshal fails
	_ = json.Unmarshal(jsonBytes, &r)

	opts := []func(*chinchon.GameState){}
	if r.MaxPoints > 0 {
		opts = append(opts, chinchon.WithMaxPoints(r.MaxPoints))
	}
	state = chinchon.New(opts...)

	bot = (newbot.New())

	nbs, err := json.Marshal(state.ToClientGameState(0))
	if err != nil {
		panic(err)
	}

	buffer := js.Global().Get("Uint8Array").New(len(nbs))
	js.CopyBytesToJS(buffer, nbs)
	return buffer
}

func chinchonRunAction(this js.Value, p []js.Value) interface{} {
	jsonBytes := make([]byte, p[0].Length())
	js.CopyBytesToGo(jsonBytes, p[0])

	newBytes := _runAction(jsonBytes)

	buffer := js.Global().Get("Uint8Array").New(len(newBytes))
	js.CopyBytesToJS(buffer, newBytes)
	return buffer
}

func chinchonBotRunAction(this js.Value, p []js.Value) interface{} {
	if !state.IsGameEnded {
		action := bot.ChooseAction(state.ToClientGameState(1))
		// fmt.Println("Action chosen by bot:", action)

		err := state.RunAction(action)
		if err != nil {
			panic(fmt.Errorf("running action: %w", err))
		}
	}

	nbs, err := json.Marshal(state.ToClientGameState(0))
	if err != nil {
		panic(fmt.Errorf("marshalling game state: %w", err))
	}

	buffer := js.Global().Get("Uint8Array").New(len(nbs))
	js.CopyBytesToJS(buffer, nbs)
	return buffer
}

func _runAction(bs []byte) []byte {
	action, err := chinchon.DeserializeAction(bs)
	if err != nil {
		panic(err)
	}
	err = state.RunAction(action)
	if err != nil {
		panic(err)
	}
	nbs, err := json.Marshal(state.ToClientGameState(0))
	if err != nil {
		panic(err)
	}
	return nbs
}
