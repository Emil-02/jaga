package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Emil-02/jaga/internal/config"
	"github.com/Emil-02/jaga/internal/fix"
	"github.com/Emil-02/jaga/internal/report"
	"github.com/Emil-02/jaga/internal/runner"
	"github.com/Emil-02/jaga/internal/score"
)

const (
	version         = "0.1.0"
	protocolVersion = "2025-06-18"
	maxFindings     = 100
)

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type server struct {
	root    string
	timeout time.Duration
}

type toolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type pathArgs struct {
	Path string `json:"path"`
}

type checkArgs struct {
	Path   string   `json:"path"`
	Checks []string `json:"checks"`
}

type gateArgs struct {
	Path   string `json:"path"`
	FailOn string `json:"fail_on"`
}

type fixArgs struct {
	Path   string `json:"path"`
	Apply  bool   `json:"apply"`
	Format string `json:"format"`
}

type findingJSON struct {
	Check    string `json:"check"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

type checkOutput struct {
	Root      string         `json:"root"`
	Score     int            `json:"score"`
	Grade     string         `json:"grade"`
	Allowed   bool           `json:"allowed"`
	FailOn    string         `json:"fail_on"`
	Summary   map[string]int `json:"summary"`
	Findings  []findingJSON  `json:"findings"`
	Truncated bool           `json:"truncated,omitempty"`
}

type gateOutput struct {
	Root     string        `json:"root"`
	Allowed  bool          `json:"allowed"`
	Score    int           `json:"score"`
	FailOn   string        `json:"fail_on"`
	Blocking []findingJSON `json:"blocking"`
	Message  string        `json:"message"`
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "jaga-mcp:", err)
		os.Exit(1)
	}
	fs := flag.NewFlagSet("jaga-mcp", flag.ContinueOnError)
	rootFlag := fs.String("root", cwd, "root direktori yang boleh dipindai")
	timeoutFlag := fs.Duration("timeout", 60*time.Second, "batas waktu per pemindaian")
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	root, err := filepath.Abs(*rootFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "jaga-mcp:", err)
		os.Exit(1)
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		fmt.Fprintln(os.Stderr, "jaga-mcp: bukan direktori:", *rootFlag)
		os.Exit(1)
	}
	srv := &server{root: root, timeout: *timeoutFlag}
	if err := srv.serve(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "jaga-mcp:", err)
		os.Exit(1)
	}
}

func (s *server) serve(in io.Reader, out io.Writer) error {
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			resp := &rpcResponse{JSONRPC: "2.0", ID: json.RawMessage("null"), Error: &rpcError{Code: -32700, Message: "parse error"}}
			writeResponse(out, &mu, &wg, resp)
			continue
		}
		if resp := s.handle(req); resp != nil {
			writeResponse(out, &mu, &wg, resp)
		}
	}
	wg.Wait()
	return sc.Err()
}

func writeResponse(out io.Writer, mu *sync.Mutex, wg *sync.WaitGroup, resp *rpcResponse) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		mu.Lock()
		defer mu.Unlock()
		enc := json.NewEncoder(out)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(resp)
	}()
}

func (s *server) handle(req rpcRequest) *rpcResponse {
	if req.JSONRPC != "2.0" {
		return rpcErr(req.ID, -32600, "invalid request: jsonrpc harus 2.0")
	}
	if len(req.ID) == 0 || string(req.ID) == "null" {
		return nil
	}
	switch req.Method {
	case "initialize":
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(req.Params, &p)
		v := p.ProtocolVersion
		if v == "" {
			v = protocolVersion
		}
		return rpcResult(req.ID, map[string]any{
			"protocolVersion": v,
			"capabilities":    map[string]any{"tools": map[string]any{"listChanged": false}},
			"serverInfo":      map[string]any{"name": "jaga-mcp", "version": version},
			"instructions":    "Gunakan jaga_check untuk memindai repo, jaga_gate sebelum commit/deploy, dan jaga_badge untuk SVG skor.",
		})
	case "ping":
		return rpcResult(req.ID, map[string]any{})
	case "tools/list":
		return rpcResult(req.ID, map[string]any{"tools": tools()})
	case "tools/call":
		var p struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return rpcErr(req.ID, -32602, "params tools/call tidak valid")
		}
		return rpcResult(req.ID, s.callTool(p.Name, p.Arguments))
	default:
		return rpcErr(req.ID, -32601, "method tidak dikenal: "+req.Method)
	}
}

func rpcResult(id json.RawMessage, result any) *rpcResponse {
	return &rpcResponse{JSONRPC: "2.0", ID: id, Result: result}
}

func rpcErr(id json.RawMessage, code int, msg string) *rpcResponse {
	return &rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}}
}

func tools() []toolDef {
	pathProp := map[string]any{
		"type":        "string",
		"description": "Path direktori yang akan dipindai, relatif terhadap root server (default: root).",
	}
	return []toolDef{
		{
			Name:        "jaga_check",
			Description: "Periksa secret, data pribadi, dan hygiene repo. Mengembalikan skor 0-100, grade, dan temuan yang sudah di-masking.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": pathProp,
					"checks": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string", "enum": runner.Names()},
						"description": "Batasi pemeriksaan (default: semua).",
					},
				},
				"additionalProperties": false,
			},
		},
		{
			Name:        "jaga_gate",
			Description: "Gerbang sebelum commit/deploy. Mengembalikan allowed=true hanya bila tidak ada temuan pada/melebihi ambang fail_on.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": pathProp,
					"fail_on": map[string]any{
						"type":        "string",
						"enum":        []string{"low", "medium", "high", "critical"},
						"description": "Ambang kegagalan; mengabaikan fail_on dari .jaga.yml.",
					},
				},
				"additionalProperties": false,
			},
		},
		{
			Name:        "jaga_badge",
			Description: "Buat badge SVG (skor/100) untuk README dari hasil pemeriksaan direktori.",
			InputSchema: map[string]any{
				"type":                 "object",
				"properties":           map[string]any{"path": pathProp},
				"additionalProperties": false,
			},
		},
		{
			Name:        "jaga_fix",
			Description: "Rencana perbaikan dari temuan jaga. Secara default dry-run: hanya file konfigurasi aman yang boleh ditulis lewat apply=true.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":   pathProp,
					"apply":  map[string]any{"type": "boolean", "description": "Tulis perbaikan aman (default: false)."},
					"format": map[string]any{"type": "string", "enum": []string{"json", "text"}, "description": "Format keluaran (default: json)."},
				},
				"additionalProperties": false,
			},
		},
	}
}

func (s *server) callTool(name string, raw json.RawMessage) map[string]any {
	switch name {
	case "jaga_check":
		var args checkArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return toolError("argumen tidak valid: " + err.Error())
		}
		payload, cfg, err := s.scan(args.Path, args.Checks)
		if err != nil {
			return toolError(err.Error())
		}
		findings, truncated := collectFindings(payload, maxFindings)
		out := checkOutput{
			Root:      payload.Root,
			Score:     payload.Score,
			Grade:     grade(payload.Score),
			Allowed:   !score.Breached(payload.Results, cfg.FailOn),
			FailOn:    string(cfg.FailOn),
			Summary:   map[string]int{},
			Findings:  findings,
			Truncated: truncated,
		}
		for _, r := range payload.Results {
			out.Summary[r.Name] = len(r.Findings)
			if r.Error != "" {
				out.Summary[r.Name+"_error"] = 1
			}
		}
		return toolJSON(out)
	case "jaga_gate":
		var args gateArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return toolError("argumen tidak valid: " + err.Error())
		}
		failOn := config.Severity(strings.ToLower(strings.TrimSpace(args.FailOn)))
		override := false
		if args.FailOn != "" {
			sev, ok := config.ParseSeverity(args.FailOn)
			if !ok {
				return toolError("fail_on tidak valid: " + args.FailOn)
			}
			failOn, override = sev, true
		}
		payload, cfg, err := s.scan(args.Path, nil)
		if err != nil {
			return toolError(err.Error())
		}
		if override {
			cfg.FailOn = failOn
		}
		blocking, truncated := collectFindings(payload, maxFindings, failOn)
		allowed := !score.Breached(payload.Results, cfg.FailOn)
		msg := "Aman untuk commit/deploy."
		if !allowed {
			msg = "Diblokir: ada temuan pada/melebihi " + string(cfg.FailOn) + "."
		}
		return toolJSON(gateOutput{
			Root:     payload.Root,
			Allowed:  allowed,
			Score:    payload.Score,
			FailOn:   string(cfg.FailOn),
			Blocking: blocking,
			Message:  msg,
		}, truncated)
	case "jaga_badge":
		var args pathArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return toolError("argumen tidak valid: " + err.Error())
		}
		payload, _, err := s.scan(args.Path, nil)
		if err != nil {
			return toolError(err.Error())
		}
		var buf bytes.Buffer
		if err := report.Render(&buf, payload, "badge"); err != nil {
			return toolError(err.Error())
		}
		return toolText(buf.String())
	case "jaga_fix":
		var args fixArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return toolError("argumen tidak valid: " + err.Error())
		}
		root, err := s.resolve(args.Path)
		if err != nil {
			return toolError(err.Error())
		}
		payload, _, err := s.scan(args.Path, nil)
		if err != nil {
			return toolError(err.Error())
		}
		payload.Root = root
		plan := fix.Build(root, payload, "")
		if args.Apply {
			if _, err := fix.Apply(root, &plan); err != nil {
				return toolError(err.Error())
			}
		}
		format := args.Format
		if format == "" {
			format = "json"
		}
		var buf bytes.Buffer
		if err := fix.Render(&buf, plan, format); err != nil {
			return toolError(err.Error())
		}
		return toolText(buf.String())
	default:
		return toolError("tool tidak dikenal: " + name)
	}
}

func (s *server) scan(pathArg string, checks []string) (report.Payload, config.Config, error) {
	root, err := s.resolve(pathArg)
	if err != nil {
		return report.Payload{}, config.Config{}, err
	}
	cfg, err := config.Load(config.Find(root))
	if err != nil {
		return report.Payload{}, config.Config{}, err
	}
	r := runner.NewDefault()
	if len(checks) > 0 {
		for _, name := range checks {
			if !runner.Valid(name) {
				return report.Payload{}, config.Config{}, fmt.Errorf("check tidak dikenal: %s", name)
			}
		}
		r = runner.New(checks...)
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	results := r.RunAll(ctx, root, cfg)
	return report.Payload{Root: root, Score: score.Calculate(results), Results: results}, cfg, nil
}

func (s *server) resolve(arg string) (string, error) {
	if arg == "" {
		arg = "."
	}
	if !filepath.IsAbs(arg) {
		arg = filepath.Join(s.root, arg)
	}
	abs, err := filepath.Abs(arg)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(s.root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path di luar root yang diizinkan: %s", s.root)
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("bukan direktori: %s", arg)
	}
	return abs, nil
}

func collectFindings(p report.Payload, limit int, minSeverity ...config.Severity) ([]findingJSON, bool) {
	out := []findingJSON{}
	min := config.SeverityLow
	if len(minSeverity) > 0 {
		min = minSeverity[0]
	}
	truncated := false
	for _, r := range p.Results {
		if len(out) >= limit {
			truncated = true
			break
		}
		if min != config.SeverityLow && r.Severity.Rank() < min.Rank() {
			continue
		}
		for _, f := range r.Findings {
			if len(out) >= limit {
				truncated = true
				break
			}
			out = append(out, findingJSON{
				Check:    r.Name,
				File:     f.File,
				Line:     f.Line,
				Message:  f.Message,
				Severity: string(r.Severity),
			})
		}
	}
	return out, truncated
}

func grade(n int) string {
	switch {
	case n >= 90:
		return "A"
	case n >= 80:
		return "B"
	case n >= 60:
		return "C"
	case n >= 40:
		return "D"
	default:
		return "F"
	}
}

func toolJSON(v any, extra ...bool) map[string]any {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return toolError(err.Error())
	}
	text := string(b)
	if len(extra) > 0 && extra[0] {
		text += "\n(catatan: temuan dipotong pada " + fmt.Sprint(maxFindings) + " item)"
	}
	return toolText(text)
}

func toolText(text string) map[string]any {
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": text}},
	}
}

func toolError(msg string) map[string]any {
	result := toolText(msg)
	result["isError"] = true
	return result
}
