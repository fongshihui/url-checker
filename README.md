r# url-checker

A small concurrency demo that checks many URLs in parallel using a worker pool.

## Run

```bash
go run . -workers 5 -timeout 3s https://example.com https://golang.org
```

Or from a file:

```bash
go run . -workers 5 -timeout 3s -in urls.txt
```

`urls.txt` example:

```
# one URL per line
https://example.com
https://golang.org
```

## Flags

- `-workers`: number of concurrent workers
- `-timeout`: per-request timeout
- `-retries`: retry count per URL
- `-in`: input file path (optional)

## Web UI

Run the server:

```bash
go run . -serve -addr :8080
```

Then open `http://localhost:8080` for the frontend and `POST /api/check` for the JSON API.
