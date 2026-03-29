# lline

`lline` is a small Go utility for counting lines in files under a directory tree.

It can:

- scan a root directory recursively
- count lines only for selected file extensions
- skip files whose paths contain excluded substrings
- process files concurrently
- limit the maximum `bufio.Scanner` token size

## Requirements

- Go 1.26+

## Install

```bash
go build -o lline ./cmd
```

Or run it directly:

```bash
go run ./cmd -extension go
```

## CLI Usage

```bash
lline -extension go
```

Available flags:

- `-extension` required, comma-separated file extensions to count, for example `go` or `go,txt`
- `-root` root directory to scan, defaults to the current working directory
- `-workers` number of concurrent workers, defaults to `GOMAXPROCS(0)`
- `-exclude` comma-separated substrings of file paths to skip
- `-buf-size` maximum scanner buffer size in MB, defaults to `1`

Examples:

```bash
lline -extension go -root .
lline -extension go,md -exclude vendor,node_modules,.git
lline -extension json -buf-size 16
```

## Output

The program prints a single total line count to `stdout`.

Large numbers are formatted with English digit grouping, for example:

```text
12,345
```

## Buffer Size

Internally, file reading uses `bufio.Scanner`, which has a maximum token size.

If a file contains very long lines, the default buffer may be too small and you may see an error like:

```text
count lines: bufio.Scanner: token too long
```

Increase the limit with `-buf-size`:

```bash
lline -extension json -buf-size 32
```

## Tests

Run tests with:

```bash
go test ./...
```
