package main

import (
	"fmt"
)

func main() {
	tokens := Tokenize(generate_data_main())
	val := Parse(tokens).(map[string]any)
	pairs := val["pairs"]
	for i, p := range pairs.([]any) {
		x0 := p.(map[string]any)["x0"].(float64)
		y0 := p.(map[string]any)["y0"].(float64)
		x1 := p.(map[string]any)["x1"].(float64)
		y1 := p.(map[string]any)["y1"].(float64)
		fmt.Printf("%d: [%.16f, %.16f], [%.16f, %.16f]\n", i, x0, y0, x1, y1)
	}
}
