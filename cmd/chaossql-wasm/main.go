//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"syscall/js"
)

var (
	activeCancel context.CancelFunc
)

func main() {
	c := make(chan struct{})

	js.Global().Set("ChaosSQL_ValidateYAML", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return `{"valid":false,"error":"missing yaml input"}`
		}
		res := ValidateScenarioYAML(args[0].String())
		bytes, _ := json.Marshal(res)
		return string(bytes)
	}))

	js.Global().Set("ChaosSQL_RunScenario", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return `{"success":false,"error":"missing configuration"}`
		}
		configStr := args[0].String()
		var callback js.Value
		if len(args) > 1 && args[1].Type() == js.TypeFunction {
			callback = args[1]
		}

		ctx, cancel := context.WithCancel(context.Background())
		activeCancel = cancel

		go func() {
			defer func() {
				if r := recover(); r != nil && callback.Truthy() {
					callback.Invoke(js.ValueOf(map[string]any{
						"type":  "ERROR",
						"error": "internal wasm panic recovered",
					}))
				}
			}()

			progressCb := func(ev ProgressEvent) {
				if callback.Truthy() {
					bytes, _ := json.Marshal(ev)
					callback.Invoke(js.ValueOf(string(bytes)))
				}
			}

			report, err := ExecuteWasmScenario(ctx, configStr, progressCb)
			if err != nil {
				if callback.Truthy() {
					callback.Invoke(js.ValueOf(map[string]any{
						"type":  "ERROR",
						"error": err.Error(),
					}))
				}
				return
			}

			reportJSON, _ := json.Marshal(report)
			if callback.Truthy() {
				callback.Invoke(js.ValueOf(map[string]any{
					"type":   "COMPLETE",
					"report": string(reportJSON),
				}))
			}
		}()

		return js.ValueOf(true)
	}))

	js.Global().Set("ChaosSQL_Cancel", js.FuncOf(func(this js.Value, args []js.Value) any {
		if activeCancel != nil {
			activeCancel()
			activeCancel = nil
			return true
		}
		return false
	}))

	js.Global().Set("ChaosSQL_GetVersion", js.FuncOf(func(this js.Value, args []js.Value) any {
		return "1.3.0-wasm"
	}))

	<-c
}
