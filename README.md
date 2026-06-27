# rclone-encrypt-claude46-sonnet

A small CLI tool that encrypts and decrypts using the rclone encryption defaults.

Rclone uses a custom salt if no salt is provided, which this tool will use by default. A few similar tools:

- https://github.com/rclone/rclone
- https://github.com/mcolatosti/rclonedecrypt
- https://github.com/br0kenpixel/rclone-rcc
- @fyears/rclone-crypt

Rclone encryption uses:
- NaCl SecretBox (XSalsa20 + Poly1305) for the file contents.
- AES256-EME for the filenames.
- scrypt for key material.

## Installation

### Windows (Scoop)

```bash
scoop bucket add cli-claude46-sonnet-go https://github.com/llm-supermarket/cli-claude46-sonnet-go
scoop install cli-claude46-sonnet-go
```

### macOS / Linux (Homebrew)

```bash
brew tap llm-supermarket/cli-claude46-sonnet-go https://github.com/llm-supermarket/cli-claude46-sonnet-go
brew install cli-claude46-sonnet-go
```

### From source

```bash
go install github.com/llm-supermarket/cli-claude46-sonnet-go@latest
```

## Usage

```
cli-claude46-sonnet-go encrypt [flags] -i <input-file> [-o <output-file>]
cli-claude46-sonnet-go decrypt [flags] -i <input-file> [-o <output-file>]
cli-claude46-sonnet-go version
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--input-file` | `-i` | — | Path to the input file (**required**) |
| `--output-file` | `-o` | auto | Path for the output file (default: encrypted/decrypted filename) |
| `--password` | `-p` | — | Password (see security note below) |
| `--salt` | — | rclone default | Optional salt / Password2 |
| `--encoding` | `-e` | `base32` | Filename encoding: `base32` or `base64` |

### Security note on `--password`

Passing passwords on the command line exposes them in shell history, process
listings (`ps aux`), and system logs. Prefer the interactive prompt by omitting
`--password`. As a less-risky alternative, use an environment variable:

```bash
export RCLONE_CRYPT_PASSWORD=mysecretpassword
cli-claude46-sonnet-go decrypt -i encrypted-file
unset RCLONE_CRYPT_PASSWORD
```

If you do use `--password`, clear the history entry afterwards:

```bash
# bash
history -d $(history 1)

# zsh
fc -ln -1   # view last command
# then edit ~/.zsh_history manually
```

## Examples

### Encrypt a file (interactive password prompt)

```bash
cli-claude46-sonnet-go encrypt -i secret.txt
# Password: ****
# Encrypted: secret.txt -> kr9tu4e1da4u3nifdd99g9tf5o
```

### Decrypt using the base32-encoded filename

```bash
cli-claude46-sonnet-go decrypt -i kr9tu4e1da4u3nifdd99g9tf5o
# Password: ****
# Decrypted: kr9tu4e1da4u3nifdd99g9tf5o -> TEST_FILE.txt
```

### Encrypt with base64 filename encoding

```bash
cli-claude46-sonnet-go encrypt -i "TEST_FILE BASE64.txt" -e base64
# Password: ****
# Encrypted: TEST_FILE BASE64.txt -> Iyxcijgc9bp3o5Y0npW6xqUvwWNcc3MA4SadB0sR6cY
```

### Decrypt a base64-encoded filename

```bash
cli-claude46-sonnet-go decrypt -i Iyxcijgc9bp3o5Y0npW6xqUvwWNcc3MA4SadB0sR6cY -e base64
# Password: ****
# Decrypted: Iyxcijgc9bp3o5Y0npW6xqUvwWNcc3MA4SadB0sR6cY -> TEST_FILE BASE64.txt
```

### Encrypt with a custom salt

```bash
cli-claude46-sonnet-go encrypt -i report.pdf --salt "my-organisation-salt" --password hunter2
```

### Decrypt with explicit output path

```bash
cli-claude46-sonnet-go decrypt -i encrypted-blob -o output.txt --password hunter2
```

### Use an environment variable for the password

```bash
export RCLONE_CRYPT_PASSWORD=hunter2
cli-claude46-sonnet-go encrypt -i secret.txt
cli-claude46-sonnet-go decrypt -i kr9tu4e1da4u3nifdd99g9tf5o
unset RCLONE_CRYPT_PASSWORD
```

## Filename encoding

| Encoding | Alphabet | Padding | Notes |
|----------|----------|---------|-------|
| `base32` | `0-9a-v` (base32hex lowercase) | None | rclone default |
| `base64` | URL-safe (`A-Za-z0-9-_`) | None | shorter names for long filenames |

## Releasing

Releases are created by pushing a version tag. GitHub Actions builds cross-platform
binaries and updates the Scoop and Brew manifests automatically.

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Uninstalling

```bash
# Windows
scoop uninstall cli-claude46-sonnet-go
scoop bucket rm cli-claude46-sonnet-go

# macOS / Linux
brew uninstall cli-claude46-sonnet-go
brew untap llm-supermarket/cli-claude46-sonnet-go
```
