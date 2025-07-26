# Paymostats

Small CLI to display how tracked time in [Paymo](https://www.paymoapp.com/) is split across projects for a chosen period.

- Installable via Homebrew
- Stores your API key securely in macOS Keychain
- Interactive mode _and_ non‑interactive flags
- Table output

## Install

```bash
brew tap Ma-Kas/tap

brew install paymostats
# or explicitly as a cask
brew install --cask paymostats
```

Normally, `brew install` can install casks or formulae; `--cask` is only needed to disambiguate when both exist.

Check what’s installed:

```bash
brew info paymostats
```

## Quick start

First authenticate:

```bash
# paste your Paymo API key to validate and store it in your Keychain
paymostats login --api-key <YOUR_PAYMO_API_KEY>
```

Then run **interactively**:

```bash
paymostats
```

Or **non-interactively** with flags:

```bash
# built-in ranges
paymostats --range week
paymostats --range 2w
paymostats --range month
paymostats --range 3m
paymostats --range 6m
paymostats --range ytd
paymostats --range all

# explicit dates (YYYY-MM-DD). If --end is omitted, it defaults to now
paymostats --start 2025-07-01 --end 2025-07-25
paymostats --start 2025-07-01
```

Logout (remove stored key):

```bash
paymostats logout
```

## Usage

```bash
paymostats [flags]

Flags:
  -r, --range string   week|2w|month|3m|6m|ytd|all
  -s, --start string   start date (YYYY-MM-DD)
  -e, --end string     end date (YYYY-MM-DD)
```

Subcommands:

```bash
paymostats login [--api-key <KEY>] # validate and store/replace your API key in Keychain
paymostats logout # remove stored key
```

## Security & privacy

- Your API key is stored in the macOS Keychain (via `go-keyring`). When running the tool, macOS may ask whether to allow access. Approve to continue.
- You can also provide `PAYMOSTATS_API_KEY` as an environment variable for local development, but Keychain is recommended for regular use.

## Troubleshooting

- For general help, run `paymostats --help`
- “No API key found” - run `paymostats login --api-key <KEY>`.
- 401 Unauthorized after an app password change - run `paymostats login --api-key <NEW_KEY>` to overwrite.
- Homebrew installation tips - see the official [docs](https://docs.brew.sh/Installation).
