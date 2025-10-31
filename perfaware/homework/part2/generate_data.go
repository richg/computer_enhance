package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"unsafe"
)

func getRandomInRange(rng *rand.Rand, min float64, max float64) float64 {
	r := rng.Float64()
	return r*(max-min) + min
}

func getRandomRegion(rng *rand.Rand) (float64, float64, float64, float64) {
	xMin := 360*rng.Float64() - 180
	xMax := getRandomInRange(rng, xMin, 180)
	yMin := 180*rng.Float64() - 90
	yMax := getRandomInRange(rng, yMin, 90)
	return xMin, yMin, xMax, yMax
}

func ToByteArray(in []float64) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&in[0])), len(in)*8)
}
func FromByteArray(in []byte) []float64 {
	return unsafe.Slice((*float64)(unsafe.Pointer(&in[0])), len(in)/8)
}

func generateData(cluster bool, generateAnswers bool, numPoints int64, seed int64) string {
	fmt.Printf("Generating data: cluster=%t, generateAnswers=%t, numPoints=%d, seed=%d\n",
		cluster, generateAnswers, numPoints, seed)
	var sb strings.Builder
	sb.Grow(int(numPoints*118) + 256) // Rough estimate of size
	sb.WriteString("{\"pairs\": [\n")
	var answers []float64 = make([]float64, numPoints)
	var totalDistance float64 = 0.0
	rng := rand.New(rand.NewSource(seed))
	var batchSize int64
	var currentRegionXMin, currentRegionYMin, currentRegionXMax, currentRegionYMax float64
	if cluster {
		currentRegionXMin, currentRegionYMin, currentRegionXMax, currentRegionYMax = getRandomRegion(rng)
		fmt.Printf("Initial Batch: X in [%.4f, %.4f], Y in [%.4f, %.4f]\n",
			currentRegionXMin, currentRegionXMax, currentRegionYMin, currentRegionYMax)
		numRegions := 64
		batchSize = numPoints/int64(numRegions) + 1
	} else {
		currentRegionXMin, currentRegionYMin, currentRegionXMax, currentRegionYMax = -180.0, -90.0, 180.0, 90.0
		fmt.Printf("Initial Batch: X in [%.4f, %.4f], Y in [%.4f, %.4f]\n",
			currentRegionXMin, currentRegionXMax, currentRegionYMin, currentRegionYMax)
		batchSize = numPoints
	}
	fmt.Printf("Generating %d points with batch size %d\n", numPoints, batchSize)
	remainingInBatch := batchSize
	fmt.Printf("Remaining in batch: %d\n", remainingInBatch)
	batchCount := 1
	for i := 0; i < int(numPoints); i++ {
		if remainingInBatch == 0 {
			currentRegionXMin, currentRegionYMin, currentRegionXMax, currentRegionYMax = getRandomRegion(rng)
			batchCount++
			// fmt.Printf("New Batch %d: X in [%.4f, %.4f], Y in [%.4f, %.4f]\n",
			// 	batchCount, currentRegionXMin, currentRegionXMax, currentRegionYMin, currentRegionYMax)
			remainingInBatch = batchSize
		}
		x0 := getRandomInRange(rng, currentRegionXMin, currentRegionXMax)
		y0 := getRandomInRange(rng, currentRegionYMin, currentRegionYMax)
		x1 := getRandomInRange(rng, currentRegionXMin, currentRegionXMax)
		y1 := getRandomInRange(rng, currentRegionYMin, currentRegionYMax)
		sb.WriteString(fmt.Sprintf("  {\"x0\": %.16f, \"y0\": %.16f, \"x1\": %.16f, \"y1\": %.16f}", x0, y0, x1, y1))
		if i != int(numPoints-1) {
			sb.WriteString(",\n")
		}
		answer := ReferenceHaversine(x0, y0, x1, y1, 6371.0)
		answers[i] = answer
		totalDistance += answer
		remainingInBatch--
	}
	sb.WriteString("\n]}")
	fn := fmt.Sprintf("geo-data-%d.json", numPoints)
	f, err := os.Create(fn)
	if err != nil {
		panic("Failed to write file")
	}
	defer f.Close()
	f.WriteString(sb.String())
	answer_fn := fmt.Sprintf("answers-%d.json", numPoints)
	answer_f, err := os.Create(answer_fn)
	if err != nil {
		panic("Failed to write file")
	}
	defer answer_f.Close()
	answer_f.Write(ToByteArray(answers))

	averageDistance := totalDistance / float64(numPoints)
	fmt.Printf("Seed: %d, cluster: %t - %.16f\n", seed, cluster, averageDistance)
	return fn
}
