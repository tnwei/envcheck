# envcheck

Verify that your `.env` lines up with your `.env.example`

## Why

Was losing my mind trying to keep env vars configured properly for deployments across multiple greenfield projects. Built this.

## Installation

Install with golang: `go get -u https://github.com/tnwei/envcheck`

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

Examples of `.env` and `.env.example`:

```
# .env
API_KEY=123456
TABLES=transactions

# .env.example
API_KEY=
TABLES=
TIMEOUT=
```