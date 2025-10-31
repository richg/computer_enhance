package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

func main() {
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	cluster := generateCmd.Bool("cluster", false, "Cluster mode")
	generateAnswers := generateCmd.Bool("answers", false, "Generate answer files")
	calculateCmd := flag.NewFlagSet("calculate", flag.ExitOnError)
	if len(os.Args) < 2 {
		fmt.Println("expected 'generate' or 'calculate' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		generateCmd.Parse(os.Args[2:])
		posArgs := generateCmd.Args()
		if len(posArgs) != 2 {
			fmt.Println("Usage: part2 generate [-cluster] [-answer] [seed] [numPoints]")
			return
		}
		seed, err := strconv.ParseInt(posArgs[0], 0, 64)
		if err != nil {
			fmt.Printf("Error parsing seed: %v\n", err)
			return
		}

		numPoints, err := strconv.ParseInt(posArgs[1], 0, 64)
		if err != nil {
			fmt.Printf("Error parsing numPoints: %v\n", err)
			return
		}
		generateData(*cluster, *generateAnswers, numPoints, seed)
	case "calculate":
		calculateCmd.Parse(os.Args[2:])
		posArgs := calculateCmd.Args()
		if len(posArgs) != 2 {
			fmt.Println("Usage: part2 calculate [dataFile] [answerFile]")
			return
		}
		dataFile := posArgs[0]
		answerFile := posArgs[1]
		LoadAndCalculate(dataFile, answerFile)
	default:
		fmt.Println("expected 'generate' or 'calculate' subcommands")
		os.Exit(1)
	}
}
