package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

func Square(x float64) float64 {
	return x * x
}

func RadiansFromDegrees(Degrees float64) float64 {
	return 0.01745329251994329577 * Degrees
}

func ReferenceHaversine(X0 float64, Y0 float64, X1 float64, Y1 float64, EarthRadius float64) float64 {
	/* NOTE(casey): This is not meant to be a "good" way to calculate the Haversine distance.
	   Instead, it attempts to follow, as closely as possible, the formula used in the real-world
	   question on which these homework exercises are loosely based.
	*/
	lat1 := Y0
	lat2 := Y1
	lon1 := X0
	lon2 := X1

	dLat := RadiansFromDegrees(lat2 - lat1)
	dLon := RadiansFromDegrees(lon2 - lon1)
	lat1 = RadiansFromDegrees(lat1)
	lat2 = RadiansFromDegrees(lat2)

	a := Square(math.Sin(dLat/2.0)) + math.Cos(lat1)*math.Cos(lat2)*Square(math.Sin(dLon/2))
	c := 2.0 * math.Asin(math.Sqrt(a))

	return EarthRadius * c
}

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

func generateData(cluster bool, numPoints int64, seed int64) string {
	var sb strings.Builder
	sb.Grow(int(numPoints*118) + 256) // Rough estimate of size
	sb.WriteString("{\"pairs\": [\n")
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
	i := numPoints
	fmt.Printf("Generating %d points with batch size %d\n", numPoints, batchSize)
	remainingInBatch := batchSize
	fmt.Printf("Remaining in batch: %d\n", remainingInBatch)
	batchCount := 1
	for i > 0 {
		if remainingInBatch == 0 {
			currentRegionXMin, currentRegionYMin, currentRegionXMax, currentRegionYMax = getRandomRegion(rng)
			batchCount++
			fmt.Printf("New Batch %d: X in [%.4f, %.4f], Y in [%.4f, %.4f]\n",
				batchCount, currentRegionXMin, currentRegionXMax, currentRegionYMin, currentRegionYMax)
			remainingInBatch = batchSize
			// sb.WriteString(fmt.Sprintf("  {\"BATCH NUM\": %d},\n", batchCount))
		}
		x0 := getRandomInRange(rng, currentRegionXMin, currentRegionXMax)
		y0 := getRandomInRange(rng, currentRegionYMin, currentRegionYMax)
		x1 := getRandomInRange(rng, currentRegionXMin, currentRegionXMax)
		y1 := getRandomInRange(rng, currentRegionYMin, currentRegionYMax)
		sb.WriteString(fmt.Sprintf("  {\"x0\": %.16f, \"y0\": %.16f, \"x1\": %.16f, \"y1\": %.16f}", x0, y0, x1, y1))
		if i > 1 {
			sb.WriteString(",\n")
		}
		totalDistance += ReferenceHaversine(x0, y0, x1, y1, 6371.0)
		remainingInBatch--
		i--
	}
	sb.WriteString("\n]}")
	fn := fmt.Sprintf("geo-data-%d.json", numPoints)
	f, err := os.Create(fmt.Sprintf("geo-data-%d.json", numPoints))
	if err != nil {
		panic("Failed to write file")
	}
	defer f.Close()
	f.WriteString(sb.String())
	averageDistance := totalDistance / float64(numPoints)
	fmt.Printf("Seed: %d, cluster: %t - %.16f\n", seed, cluster, averageDistance)
	return fn
}

func generate_data_main() string {
	cluster := flag.Bool("cluster", false, "Cluster mode")
	flag.Parse()
	posArgs := flag.Args()
	if len(posArgs) != 2 {
		fmt.Println("Usage: go run generate_data.go [-cluster] [seed] [numPoints]")
		return ""
	}
	seed, err := strconv.ParseInt(posArgs[0], 0, 64)
	if err != nil {
		fmt.Printf("Error parsing seed: %v\n", err)
		return ""
	}

	numPoints, err := strconv.ParseInt(posArgs[1], 0, 64)
	if err != nil {
		fmt.Printf("Error parsing numPoints: %v\n", err)
		return ""
	}
	return generateData(*cluster, numPoints, seed)
}
