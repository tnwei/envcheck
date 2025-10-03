# envcheck

Verify that your `.env` lines up with your `.env.example`

## Why

Was losing my mind trying to keep env vars configured properly for deployments across multiple greenfield projects. Built this.

## Installation

Install with golang: `go install github.com/tnwei/envcheck@latest`

Or `git clone` this repo, then run `go build` in the repo directory

Drop the resultant binary into somewhere in your `PATH`. I recommend `~/.local/bin`.

## Usage

Run:

```bash
$ envcheck
Usage:
  envcheck [command] [options]

Commands:
  envcheck list [dir] <env_file> <example_file>     - List env files and difference
  envcheck create <env_file> <example_file>         - Create env file from example
  envcheck update <env_file> <example_file>         - Update env file with missing keys

Flags:
  --env_file defaults to ".env"
  --example_file defaults to ".env.example"

Examples:
  envcheck list
  envcheck list ./deploy/
  envcheck create .env   
  envcheck create prod/.env prod/.env.example
  envcheck create .env.staging .env.example
  envcheck update .env
  envcheck update .env.development dev/.env.example
```

## Sample output

Given `.env.example`:

```
API_KEY=123456
TABLES=transactions
TIMEOUT=
```

```
$ envcheck list
Listing env files in path: .
Found 1 example files and 0 env files.

⚠ .env doesn't exist (template available: .env.example)

$ envcheck create
✓ Created .env with 2 keys from .env.example

$ cat .env
API_KEY=123456
TABLES=transactions
TIMEOUT=

$ echo PORT=4321 >> .env.example

$ envcheck update
✓ Added 1 missing keys to .env
  + PORT

$ envcheck update
✓ .env is in sync with .env.example

$ envcheck create
Error: ✗ Error: .env already exists. Use 'update' instead.
Usage:
  envcheck create [flags]

Flags:
  -e, --env-file string       Path to the environment file (default ".env")
  -x, --example-file string   Path to the example file (default ".env.example")
  -h, --help                  help for create

✗ Error: .env already exists. Use 'update' instead.

$ envcheck list
Listing env files in path: .
Found 1 example files and 1 env files.

✓ .env is in sync with .env.example
```
