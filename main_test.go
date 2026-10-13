package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const testHeader = "value,income,age,rooms,bedrooms,pop,hh\n"

// TestConvertValidCSV checks the row count, JSON validity, and numeric values.
func TestConvertValidCSV(t *testing.T) {
	input := testHeader +
		"452600,8.3252,41,880,129,322,126\n" +
		"358500,8.3014,21,7099,1106,2401,1138\n"

	var output bytes.Buffer

	count, err := convert(strings.NewReader(input), &output)
	if err != nil {
		t.Fatalf("convert returned an error: %v", err)
	}

	if count != 2 {
		t.Fatalf("expected 2 rows, got %d", count)
	}

	if !strings.HasSuffix(output.String(), "\n") {
		t.Error("output must end with a newline")
	}

	lines := strings.Split(strings.TrimSuffix(output.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 JSON lines, got %d", len(lines))
	}

	expected := []map[string]float64{
		{
			"value": 452600, "income": 8.3252, "age": 41,
			"rooms": 880, "bedrooms": 129, "pop": 322, "hh": 126,
		},
		{
			"value": 358500, "income": 8.3014, "age": 21,
			"rooms": 7099, "bedrooms": 1106, "pop": 2401, "hh": 1138,
		},
	}

	for i, line := range lines {
		var actual map[string]float64

		if err := json.Unmarshal([]byte(line), &actual); err != nil {
			t.Fatalf("line %d is not valid numeric JSON: %v", i+1, err)
		}

		if len(actual) != len(expected[i]) {
			t.Fatalf("line %d has the wrong number of fields", i+1)
		}

		for field, want := range expected[i] {
			got, exists := actual[field]
			if !exists || got != want {
				t.Errorf(
					"line %d, field %q: got %v, want %v",
					i+1, field, got, want,
				)
			}
		}
	}
}

// TestConvertInvalidCSV checks several common input problems.
func TestConvertInvalidCSV(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty file",
			input: "",
		},
		{
			name:  "wrong column count",
			input: "value,income\n452600,8.3252\n",
		},
		{
			name: "incorrect header",
			input: "price,income,age,rooms,bedrooms,pop,hh\n" +
				"452600,8.3252,41,880,129,322,126\n",
		},
		{
			name: "missing value",
			input: testHeader +
				"452600,8.3252,41,880,129,322\n",
		},
		{
			name: "non-numeric value",
			input: testHeader +
				"452600,invalid,41,880,129,322,126\n",
		},
		{
			name: "invalid JSON number",
			input: testHeader +
				"452600,NaN,41,880,129,322,126\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer

			_, err := convert(strings.NewReader(test.input), &output)
			if err == nil {
				t.Fatal("expected an error, got none")
			}
		})
	}
}

// TestRunInvalidArguments checks the command-line usage validation.
func TestRunInvalidArguments(t *testing.T) {
	tests := [][]string{
		nil,
		{"input.csv"},
		{"input.csv", "output.jl", "extra"},
	}

	for _, args := range tests {
		if err := run(args); err == nil {
			t.Errorf("expected an error for arguments %v", args)
		}
	}
}

// TestRunSuccess verifies conversion using actual temporary files.
func TestRunSuccess(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.csv")
	outputPath := filepath.Join(dir, "output.jl")

	input := testHeader +
		"452600,8.3252,41,880,129,322,126\n"

	if err := os.WriteFile(inputPath, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{inputPath, outputPath}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	if !bytes.HasSuffix(data, []byte("\n")) {
		t.Error("output must end with a newline")
	}

	var house map[string]float64
	if err := json.Unmarshal(bytes.TrimSpace(data), &house); err != nil {
		t.Fatalf("output is not valid numeric JSON: %v", err)
	}

	if len(house) != 7 || house["value"] != 452600 {
		t.Errorf("unexpected output: %v", house)
	}
}

// TestRunMissingInput verifies that a missing input file is rejected.
func TestRunMissingInput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "missing.csv")
	outputPath := filepath.Join(dir, "output.jl")

	if err := run([]string{inputPath, outputPath}); err == nil {
		t.Fatal("expected an error for a missing input file")
	}

	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("output should not exist; got error: %v", err)
	}
}

// TestRunExistingOutput verifies that existing output is preserved.
func TestRunExistingOutput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.csv")
	outputPath := filepath.Join(dir, "output.jl")

	input := testHeader +
		"452600,8.3252,41,880,129,322,126\n"
	original := []byte("Keep this existing file unchanged.\n")

	if err := os.WriteFile(inputPath, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outputPath, original, 0644); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{inputPath, outputPath}); err == nil {
		t.Fatal("expected an error for an existing output file")
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, original) {
		t.Fatal("existing output file was changed")
	}
}

// TestRunInvalidInputRemovesOutput verifies cleanup after partial conversion.
func TestRunInvalidInputRemovesOutput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.csv")
	outputPath := filepath.Join(dir, "output.jl")

	input := testHeader +
		"452600,8.3252,41,880,129,322,126\n" +
		"358500,invalid,21,7099,1106,2401,1138\n"

	if err := os.WriteFile(inputPath, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{inputPath, outputPath}); err == nil {
		t.Fatal("expected an error for invalid input")
	}

	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("incomplete output should be removed; got error: %v", err)
	}
}

// BenchmarkConvert measures conversion of 1,000 synthetic housing records.
// Output is discarded, so this measures conversion rather than disk writes.
func BenchmarkConvert(b *testing.B) {
	const rows = 1000

	input := testHeader + strings.Repeat(
		"452600,8.3252,41,880,129,322,126\n",
		rows,
	)

	b.ReportAllocs()
	b.SetBytes(int64(len(input)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		count, err := convert(strings.NewReader(input), io.Discard)
		if err != nil {
			b.Fatal(err)
		}
		if count != rows {
			b.Fatalf("expected %d rows, got %d", rows, count)
		}
	}
}

// TestHousingOutput validates the actual generated file against the input CSV.
func TestHousingOutput(t *testing.T) {
	if os.Getenv("CHECK_HOUSING_OUTPUT") != "1" {
		t.Skip("set CHECK_HOUSING_OUTPUT=1 to validate the housing files")
	}

	input, err := os.Open("housesInput.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()

	output, err := os.Open("housesOutput.jl")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()

	reader := csv.NewReader(input)
	headers, err := reader.Read()
	if err != nil {
		t.Fatalf("read CSV header: %v", err)
	}

	scanner := bufio.NewScanner(output)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	count := 0

	for scanner.Scan() {
		count++

		var house map[string]float64
		if err := json.Unmarshal(scanner.Bytes(), &house); err != nil {
			t.Fatalf("output line %d is invalid numeric JSON: %v", count, err)
		}

		record, err := reader.Read()
		if err != nil {
			t.Fatalf("read CSV record for output line %d: %v", count, err)
		}

		if len(house) != len(headers) {
			t.Fatalf("output line %d has an incorrect field count", count)
		}

		for i, field := range headers {
			want, err := strconv.ParseFloat(record[i], 64)
			if err != nil {
				t.Fatalf("CSV record %d, field %q: %v", count, field, err)
			}

			got, exists := house[field]
			if !exists || got != want {
				t.Fatalf(
					"output line %d, field %q: got %v, want %v",
					count, field, got, want,
				)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("read output: %v", err)
	}

	if _, err := reader.Read(); err != io.EOF {
		t.Fatalf("CSV has remaining records or a read error: %v", err)
	}

	if count == 0 {
		t.Fatal("output contains no data")
	}

	t.Logf(
		"Validated %d JSON lines; all fields match the CSV records.",
		count,
	)
}
