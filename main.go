package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/llm-supermarket/cli-claude46-sonnet-go/internal/crypt"
	"golang.org/x/term"
)

var version = "dev"

const passwordWarning = `WARNING: Passing passwords on the command line is insecure.
  - Your password may appear in shell history, process listings, and logs.
  - Prefer using the interactive prompt (omit --password).
  - If you must script this, use an environment variable:
      export RCLONE_CRYPT_PASSWORD=<password>
      cli-claude46-sonnet-go decrypt -i <file>
  - Clear your shell history afterwards: history -d $(history 1)
`

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "encrypt":
		if err := runEncryptCmd(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "decrypt":
		if err := runDecryptCmd(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "version", "--version", "-v":
		fmt.Printf("cli-claude46-sonnet-go version %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`cli-claude46-sonnet-go - rclone-compatible file encryption/decryption

Usage:
  cli-claude46-sonnet-go encrypt [flags] -i <input-file> [-o <output-file>]
  cli-claude46-sonnet-go decrypt [flags] -i <input-file> [-o <output-file>]
  cli-claude46-sonnet-go version

Flags:
  -i, --input-file   Path to input file (required)
  -o, --output-file  Path to output file (default: encrypted/decrypted filename)
  -p, --password     Password (see WARNING below; prefer interactive prompt)
      --salt         Optional salt (default: rclone built-in salt)
  -e, --encoding     Filename encoding: base32 (default) or base64

WARNING: Using --password on the command line exposes your password in shell
history and process listings. Omit it to be prompted securely instead.

Examples:
  # Encrypt a file (prompted for password)
  cli-claude46-sonnet-go encrypt -i secret.txt

  # Decrypt using base64 filename encoding
  cli-claude46-sonnet-go decrypt -i Iyxcijgc9bp3o5Y0npW6xqUvwWNcc3MA4SadB0sR6cY -e base64

  # Encrypt with a custom salt (prompted for password)
  cli-claude46-sonnet-go encrypt -i report.pdf --salt mysalt

  # Decrypt specifying password via env var (less insecure than --password)
  RCLONE_CRYPT_PASSWORD=hunter2 cli-claude46-sonnet-go decrypt -i <encrypted-file>
`)
}

type commonFlags struct {
	inputFile  string
	outputFile string
	password   string
	salt       string
	encoding   string
}

func parseCommonFlags(args []string) (*commonFlags, error) {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	f := &commonFlags{}

	fs.StringVar(&f.inputFile, "i", "", "")
	fs.StringVar(&f.inputFile, "input-file", "", "")
	fs.StringVar(&f.outputFile, "o", "", "")
	fs.StringVar(&f.outputFile, "output-file", "", "")
	fs.StringVar(&f.password, "p", "", "")
	fs.StringVar(&f.password, "password", "", "")
	fs.StringVar(&f.salt, "salt", "", "")
	fs.StringVar(&f.encoding, "e", "base32", "")
	fs.StringVar(&f.encoding, "encoding", "base32", "")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if f.inputFile == "" {
		return nil, fmt.Errorf("--input-file / -i is required")
	}
	return f, nil
}

func resolvePassword(f *commonFlags) (string, error) {
	if f.password != "" {
		fmt.Fprint(os.Stderr, passwordWarning)
		return f.password, nil
	}
	if p := os.Getenv("RCLONE_CRYPT_PASSWORD"); p != "" {
		return p, nil
	}
	return promptSecret("Password: ")
}

func promptSecret(label string) (string, error) {
	fmt.Fprint(os.Stderr, label)
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("reading password: %w", err)
	}
	return string(pw), nil
}

func runEncryptCmd(args []string) error {
	f, err := parseCommonFlags(args)
	if err != nil {
		return err
	}

	password, err := resolvePassword(f)
	if err != nil {
		return err
	}

	enc, err := crypt.ParseEncoding(f.encoding)
	if err != nil {
		return err
	}

	c, err := crypt.New(password, f.salt, enc)
	if err != nil {
		return fmt.Errorf("initialising cipher: %w", err)
	}

	outputFile := f.outputFile
	if outputFile == "" {
		base := filepath.Base(f.inputFile)
		outputFile, err = c.EncryptFileName(base)
		if err != nil {
			return fmt.Errorf("encrypting filename: %w", err)
		}
	}

	in, err := os.Open(f.inputFile)
	if err != nil {
		return fmt.Errorf("opening input: %w", err)
	}
	defer in.Close()

	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("creating output: %w", err)
	}
	defer out.Close()

	if err := c.EncryptContent(out, in); err != nil {
		return fmt.Errorf("encrypting content: %w", err)
	}

	fmt.Printf("Encrypted: %s -> %s\n", f.inputFile, outputFile)
	return nil
}

func runDecryptCmd(args []string) error {
	f, err := parseCommonFlags(args)
	if err != nil {
		return err
	}

	password, err := resolvePassword(f)
	if err != nil {
		return err
	}

	enc, err := crypt.ParseEncoding(f.encoding)
	if err != nil {
		return err
	}

	c, err := crypt.New(password, f.salt, enc)
	if err != nil {
		return fmt.Errorf("initialising cipher: %w", err)
	}

	outputFile := f.outputFile
	if outputFile == "" {
		base := filepath.Base(f.inputFile)
		outputFile, err = c.DecryptFileName(base)
		if err != nil {
			return fmt.Errorf("decrypting filename %q: %w", base, err)
		}
	}

	in, err := os.Open(f.inputFile)
	if err != nil {
		return fmt.Errorf("opening input: %w", err)
	}
	defer in.Close()

	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("creating output: %w", err)
	}
	defer out.Close()

	if err := c.DecryptContent(out, in); err != nil {
		// Remove incomplete output on failure
		out.Close()
		os.Remove(outputFile)
		return fmt.Errorf("decrypting content: %w", err)
	}

	fmt.Printf("Decrypted: %s -> %s\n", f.inputFile, outputFile)
	return nil
}

