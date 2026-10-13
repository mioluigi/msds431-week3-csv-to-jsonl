# MSDS 431 Week 3: CSV-to-JSON-Lines Converter

## Overview

This Go command-line application converts the California housing dataset
from CSV into JSON Lines.

The application accepts an input CSV path and an output path. Each CSV
data record becomes one JSON object on a separate line in the output.
Numeric values are written as JSON numbers, not strings.

This implementation is specific to the housing dataset. It is not a
general-purpose converter for arbitrary CSV schemas.

## Requirements

- Go 1.27.1 or later, as specified in go.mod.
- The assignment dataset, housesInput.csv, downloaded from Canvas.
- A terminal opened in this project's root folder.

The application uses only Go's standard library. No third-party packages
are required.

## Project Files

| File | Purpose |
|---|---|
| go.mod | Defines the module path and Go version requirement |
| main.go | Implements argument validation and CSV-to-JSON-Lines conversion |
| main_test.go | Contains unit tests, file-handling tests, a benchmark, and an optional full-output validation test |
| housesInput.csv | California housing input dataset supplied for the assignment |
| housesOutput.jl | Generated JSON Lines output |
| README.md | Documents setup, usage, testing, design, and AI assistance |

Download housesInput.csv from the Week 3 assignment in Canvas and place
it in the project root if it is not already present.

## Input Format

The input must contain these seven columns in this order:

```csv
"value","income","age","rooms","bedrooms","pop","hh"
452600,8.3252,41,880,129,322,126
358500,8.3014,21,7099,1106,2401,1138
```

The program checks the header, record lengths, and numeric values.
Fields must also be encodable as valid JSON numbers.

## Run from Source

From the project root:

```sh
go run main.go housesInput.csv housesOutput.jl
```

Expected success message for the supplied dataset:

```text
Converted 20640 rows to housesOutput.jl
```

Paths containing spaces must be enclosed in quotation marks.

For example:

```sh
go run main.go "input data/housesInput.csv" "output data/housesOutput.jl"
```

Parent directories for the output must already exist.

### Existing Output Files

The application intentionally refuses to overwrite an existing output
file. This protects existing data and prevents overwriting the input.

If housesOutput.jl already exists, use a new output filename:

```sh
go run main.go housesInput.csv housesOutput-new.jl
```

Alternatively, deliberately remove the old output before regenerating it.

### Error Handling

The application reports errors to standard error and exits with a
nonzero status when conversion fails.

Examples of rejected input include:

- An incorrect number of command-line arguments.
- A missing or unreadable input file.
- An existing output file.
- An incorrect CSV header.
- A record with an incorrect number of fields.
- Non-numeric values or invalid JSON numbers.

If conversion fails after creating the output file, the program attempts
to remove the incomplete output.

## Output Format

Each output line is a separate JSON object. For example:

```json
{"age":41,"bedrooms":129,"hh":126,"income":8.3252,"pop":322,"rooms":880,"value":452600}
```

The file is JSON Lines, not a JSON array. There are no surrounding array
brackets or commas between records.

JSON field order does not need to match CSV column order. Values are
associated with their field names.

The output filename may use .jl, .jsonl, or .ndjson; this assignment uses
housesOutput.jl.

## Design

The application separates responsibilities into three functions:

- main: reports errors and sets the process exit status.
- run: validates command-line arguments and manages input/output files.
- convert: validates the CSV schema and converts records into JSON lines.

The converter reads and writes one record at a time rather than loading
the entire dataset into memory.

## Formatting and Static Checks

Run:

```sh
go fmt ./...
go vet ./...
```

## Automated Tests

Run all ordinary tests with statement coverage:

```sh
go test -v -cover ./...
```

Tests cover:

- Successful conversion and preservation of numeric values.
- Valid JSON and newline-separated output.
- Empty input and incorrect headers.
- Missing fields and invalid numeric values.
- Incorrect command-line arguments.
- Successful conversion using temporary files.
- Missing input files.
- Protection of existing output files.
- Removal of incomplete output after invalid input.

The full-dataset validation test is skipped unless explicitly enabled.

Statement coverage measures which statements the tests execute; it does
not guarantee the absence of defects.

### Verified Test Results

The latest ordinary test run passed with 93.0% statement coverage:

```text
ok github.com/mioluigi/msds431-week3-csv-to-jsonl
coverage: 93.0% of statements
```

The optional full-dataset validation test was run separately and passed
for all 20,640 records.

## Full-Dataset Validation

First generate housesOutput.jl if it does not already exist.

The optional integration test checks every output line and compares every
field with the corresponding input CSV record. It also checks for extra
or missing records.

### macOS or Linux

```sh
CHECK_HOUSING_OUTPUT=1 go test -run '^TestHousingOutput$' -v
```

### Windows PowerShell

```powershell
$env:CHECK_HOUSING_OUTPUT = "1"
go test -run '^TestHousingOutput$' -v
Remove-Item Env:CHECK_HOUSING_OUTPUT
```

### Windows Command Prompt

```bat
set CHECK_HOUSING_OUTPUT=1
go test -run "^TestHousingOutput$" -v
set CHECK_HOUSING_OUTPUT=
```

### Verified Result

Full-dataset validation passed:

```text
Validated 20640 JSON lines; all fields match the CSV records.
```

The reported test duration was 0.04 seconds. This is the validation
test duration, not the conversion runtime.

## Performance Benchmark

Run:

```sh
go test -run '^$' -bench BenchmarkConvert -benchmem -count=3
```

Each benchmark operation converts 1,000 synthetic housing records.
Output is discarded to measure conversion without disk-writing overhead.

Measured results from three local runs:

| Run | ns/op | MB/s | B/op | allocs/op |
|---|---:|---:|---:|---:|
| 1 | 1,423,411 | 23.21 | 637,359 | 14,017 |
| 2 | 1,425,442 | 23.18 | 637,361 | 14,017 |
| 3 | 1,425,248 | 23.18 | 637,353 | 14,017 |

Conversion took approximately 1.425 milliseconds per 1,000 records.

B/op reports total bytes allocated per operation, not peak memory usage.
The benchmark does not measure executable startup, disk I/O, or
full-program peak memory. Results depend on the machine and Go version.

## Build and Run an Executable

Run the commands for your operating system from the project root.

### macOS or Linux

Build:

```sh
go build -o csv-to-jsonl .
```

Run:

```sh
./csv-to-jsonl housesInput.csv housesOutput-executable.jl
```

This builds a terminal executable, not a graphical macOS .app bundle.

### Windows PowerShell

Build:

```powershell
go build -o csv-to-jsonl.exe .
```

Run:

```powershell
.\csv-to-jsonl.exe housesInput.csv housesOutput-executable.jl
```

### Windows Command Prompt

Build:

```bat
go build -o csv-to-jsonl.exe .
```

Run:

```bat
csv-to-jsonl.exe housesInput.csv housesOutput-executable.jl
```

The example uses a new output filename because the application refuses
to overwrite existing files.

Building requires Go. The resulting executable can run on a compatible
operating system and architecture without invoking go run.

## Verified macOS Executable Results

The application was built successfully on macOS using:

```sh
go build -o csv-to-jsonl .
```

The executable generated housesOutput-executable.jl. Comparing this
file with the previously validated housesOutput.jl using diff showed
no differences.

## Dataset Source

The input dataset is housesInput.csv, provided through the MSDS 431
Week 3 assignment in Canvas.

The assignment identifies the data as derived from:

Miller, Thomas W. 2015. Modeling Techniques in Predictive Analytics with
Python and R: A Guide to Data Science. Pearson Education.

## AI Assistance Disclosure

I used Perplexity to help interpret the assignment requirements,
understand Go module setup, and work through development in VS Code.

The assistant generated draft converter code, unit tests, file-handling
tests, a benchmark, a full-output validation test, and README content.
It also explained the commands used to run and test the program.

I ran the converter and testing commands locally and supplied the
benchmark and full-dataset validation results recorded in this README.