package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Emil-02/jaga/internal/config"
	"github.com/Emil-02/jaga/internal/report"
	"github.com/Emil-02/jaga/internal/runner"
	"github.com/Emil-02/jaga/internal/score"
	"github.com/Emil-02/jaga/internal/setup"
)

const version = "0.1.0"

type options struct {
	path   string
	format string
	output string
	config string
	failOn string
	ci     bool
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "check":
		opts, err := parseCheckArgs(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, "jaga:", err)
			os.Exit(2)
		}
		os.Exit(runCheck(opts))
	case "init":
		opts, err := parseInitArgs(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, "jaga:", err)
			os.Exit(2)
		}
		os.Exit(runInit(opts))
	case "version":
		fmt.Println("jaga", version)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintln(os.Stderr, "jaga: perintah tidak dikenal:", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func parseCheckArgs(args []string) (options, error) {
	opts := options{path: ".", format: "text", failOn: ""}
	for i := 0; i < len(args); i++ {
		a := args[i]
		next := func() (string, error) {
			if i+1 >= len(args) {
				return "", fmt.Errorf("nilai untuk %s tidak ada", a)
			}
			i++
			return args[i], nil
		}
		switch a {
		case "--format", "-format":
			v, err := next()
			if err != nil {
				return opts, err
			}
			opts.format = v
		case "--output", "-o":
			v, err := next()
			if err != nil {
				return opts, err
			}
			opts.output = v
		case "--config":
			v, err := next()
			if err != nil {
				return opts, err
			}
			opts.config = v
		case "--fail-on":
			v, err := next()
			if err != nil {
				return opts, err
			}
			opts.failOn = v
		case "--ci":
			opts.ci = true
		case "-h", "--help":
			usage()
			os.Exit(0)
		default:
			if opts.path != "." {
				return opts, fmt.Errorf("argumen tidak dikenal: %s", a)
			}
			opts.path = a
		}
	}
	return opts, nil
}

func runCheck(opts options) int {
	root, err := filepath.Abs(opts.path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "jaga:", err)
		return 2
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		fmt.Fprintln(os.Stderr, "jaga: bukan direktori:", opts.path)
		return 2
	}

	cfgPath := opts.config
	if cfgPath == "" {
		cfgPath = config.Find(root)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "jaga:", err)
		return 2
	}
	if opts.failOn != "" {
		sev, ok := config.ParseSeverity(opts.failOn)
		if !ok {
			fmt.Fprintln(os.Stderr, "jaga: fail_on tidak valid:", opts.failOn)
			return 2
		}
		cfg.FailOn = sev
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	results := runner.NewDefault().RunAll(ctx, root, cfg)
	s := score.Calculate(results)

	var out io.Writer = os.Stdout
	if opts.output != "" {
		f, err := os.Create(opts.output)
		if err != nil {
			fmt.Fprintln(os.Stderr, "jaga:", err)
			return 2
		}
		defer f.Close()
		out = f
	}
	payload := report.Payload{
		Root:        root,
		Score:       s,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Results:     results,
	}
	if err := report.Render(out, payload, opts.format); err != nil {
		fmt.Fprintln(os.Stderr, "jaga:", err)
		return 2
	}
	if opts.ci && score.Breached(results, cfg.FailOn) {
		return 1
	}
	return 0
}

type initOptions struct {
	path        string
	force       bool
	skipScan    bool
	authorName  string
	authorEmail string
}

func parseInitArgs(args []string) (initOptions, error) {
	opts := initOptions{
		path:        ".",
		authorName:  os.Getenv("JAGA_AUTHOR_NAME"),
		authorEmail: os.Getenv("JAGA_AUTHOR_EMAIL"),
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		next := func() (string, error) {
			if i+1 >= len(args) {
				return "", fmt.Errorf("nilai untuk %s tidak ada", a)
			}
			i++
			return args[i], nil
		}
		switch a {
		case "--force":
			opts.force = true
		case "--skip-scan":
			opts.skipScan = true
		case "--author-name":
			v, err := next()
			if err != nil {
				return opts, err
			}
			opts.authorName = v
		case "--author-email":
			v, err := next()
			if err != nil {
				return opts, err
			}
			opts.authorEmail = v
		case "-h", "--help":
			usage()
			os.Exit(0)
		default:
			if strings.HasPrefix(a, "-") {
				return opts, fmt.Errorf("argumen tidak dikenal: %s", a)
			}
			if opts.path != "." {
				return opts, fmt.Errorf("argumen tidak dikenal: %s", a)
			}
			opts.path = a
		}
	}
	return opts, nil
}

func runInit(opts initOptions) int {
	result, err := setup.Run(setup.Options{
		Root:        opts.path,
		Force:       opts.force,
		SkipScan:    opts.skipScan,
		AuthorName:  opts.authorName,
		AuthorEmail: opts.authorEmail,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "jaga:", err)
		return 2
	}
	for _, file := range result.Created {
		fmt.Println("+", file)
	}
	for _, file := range result.Skipped {
		fmt.Println("=", file)
	}
	if len(result.Created) == 0 {
		fmt.Println("tidak ada perubahan")
	}
	return 0
}

func usage() {
	fmt.Print(`jaga — satu gerbang sebelum deploy

Pemakaian:
  jaga check [path] [opsi]
  jaga init [path] [opsi]
  jaga version

Opsi init:
  --force                  timpa file yang sudah ada
  --skip-scan              jangan buat jaga-report.json
  --author-name NAMA       identitas commit di workflow (atau JAGA_AUTHOR_NAME)
  --author-email EMAIL     identitas commit di workflow (atau JAGA_AUTHOR_EMAIL)

Opsi check:
  --format text|json|sarif|badge|html|markdown  format keluaran (default: text)
  -o, --output FILE                tulis ke file
  --config FILE                    paksa file konfigurasi
  --fail-on low|medium|high|critical  ambang kegagalan
  --ci                             keluar dengan kode 1 bila ambang terlampaui

Contoh:
  jaga check .
  jaga check --format sarif -o jaga.sarif .
  jaga check --ci --fail-on high .
  jaga init . --author-name "emil rahmat" --author-email "189325051+Emil-02@users.noreply.github.com"
`)
}
