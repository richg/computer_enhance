package main

import (
	"fmt"
	"os"
)

func CalculateReference(answers []float64) float64 {
	var total float64 = 0.0
	for _, ans := range answers {
		total += ans
	}
	average := total / float64(len(answers))
	return average
}

func Calculate(pairs []any) float64 {
	var totalDistance float64 = 0.0
	for _, p := range pairs {
		x0 := p.(map[string]any)["x0"].(float64)
		y0 := p.(map[string]any)["y0"].(float64)
		x1 := p.(map[string]any)["x1"].(float64)
		y1 := p.(map[string]any)["y1"].(float64)
		ans := ReferenceHaversine(x0, y0, x1, y1, 6371.0)
		totalDistance += ans
	}
	return totalDistance / float64(len(pairs))
}

func LoadAndCalculate(filename string, answerFilename string) {
	tokens := Tokenize(filename)
	val := Parse(tokens).(map[string]any)
	pairs := val["pairs"]
	averageDistance := Calculate(pairs.([]any))
	f, err := os.ReadFile(answerFilename)
	if err != nil {
		panic("Failed to read answer file")
	}
	answers := FromByteArray(f)
	referenceAverage := CalculateReference(answers)
	fmt.Printf("Calculated Average Distance: %.16f\n", averageDistance)
	fmt.Printf("Reference Average Distance: %.16f\n", referenceAverage)
	fmt.Printf("Difference: %.16f\n", averageDistance-referenceAverage)
}
