package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

// run validates the arguments and converts the housing CSV to JSON lines.
func run(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf(
			"usage: go run main.go <input.csv> <output.jl>",
		)
	}

	inputPath := args[0]
	outputPath := args[1]

	input, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open input file: %w", err)
	}
	defer input.Close()

	// Refuse to overwrite an existing file, including the input file.
	output, err := os.OpenFile(
		outputPath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644,
	)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}

	count, conversionErr := convert(input, output)
	closeErr := output.Close()

	if conversionErr != nil {
		// Remove incomplete output if conversion fails.
		os.Remove(outputPath)
		return conversionErr
	}
	if closeErr != nil {
		os.Remove(outputPath)
		return fmt.Errorf("close output file: %w", closeErr)
	}

	fmt.Printf("Converted %d rows to %s\n", count, outputPath)
	return nil
}

// convert reads one CSV record at a time and writes one JSON object per line.
func convert(input io.Reader, output io.Writer) (int, error) {
	reader := csv.NewReader(input)

	headers, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("read CSV header: %w", err)
	}

	expected := []string{
		"value", "income", "age", "rooms", "bedrooms", "pop", "hh",
	}

	if len(headers) != len(expected) {
		return 0, fmt.Errorf("CSV must have exactly seven columns")
	}

	for i, name := range expected {
		if headers[i] != name {
			return 0, fmt.Errorf(
				"column %d must be %q, got %q",
				i+1, name, headers[i],
			)
		}
	}

	encoder := json.NewEncoder(output)
	count := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, fmt.Errorf(
				"read data row %d: %w", count+1, err,
			)
		}

		house := make(map[string]json.Number, len(headers))

		for i, text := range record {
			if _, err := strconv.ParseFloat(text, 64); err != nil {
				return count, fmt.Errorf(
					"data row %d, column %q: invalid number %q",
					count+1, headers[i], text,
				)
			}
			house[headers[i]] = json.Number(text)
		}

		if err := encoder.Encode(house); err != nil {
			return count, fmt.Errorf(
				"write data row %d: %w", count+1, err,
			)
		}

		count++
	}

	return count, nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
