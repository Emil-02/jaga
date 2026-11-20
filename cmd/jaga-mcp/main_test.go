package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testServer(t *testing.T) (*server, string) {
	t.Helper()
	dir := t.TempDir()
	return &server{root: dir, timeout: 5 * time.Second}, dir
}

func put(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hygiene(t *testing.T, dir string) {
	t.Helper()
	put(t, dir, "README.md", "# contoh\n")
	put(t, dir, "LICENSE", "MIT\n")
	put(t, dir, ".gitignore", "node_modules\n")
	put(t, dir, ".github/workflows/ci.yml", "name: ci\n")
}

func request(method string, id string, params any) rpcRequest {
	raw, _ := json.Marshal(params)
	return rpcRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(id),
		Method:  method,
		Params:  raw,
	}
}

func textOf(t *testing.T, resp *rpcResponse) string {
	t.Helper()
	if resp.Error != nil {
		t.Fatalf("rpc error: %+v", resp.Error)
	}
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected result: %T", resp.Result)
	}
	content := result["content"].([]map[string]any)
	return content[0]["text"].(string)
}

func callTool(t *testing.T, s *server, name string, args any) (string, bool) {
	t.Helper()
	raw, _ := json.Marshal(args)
	resp := s.handle(request("tools/call", `1`, map[string]any{"name": name, "arguments": json.RawMessage(raw)}))
	text := textOf(t, resp)
	result := resp.Result.(map[string]any)
	isError, _ := result["isError"].(bool)
	return text, isError
}

func TestInitialize(t *testing.T) {
	s, _ := testServer(t)
	resp := s.handle(request("initialize", `1`, map[string]any{"protocolVersion": "2025-03-26"}))
	result := resp.Result.(map[string]any)
	if result["protocolVersion"] != "2025-03-26" {
		t.Fatalf("protocolVersion = %v", result["protocolVersion"])
	}
	caps := result["capabilities"].(map[string]any)
	if _, ok := caps["tools"]; !ok {
		t.Fatal("capabilities.tools tidak ada")
	}
	info := result["serverInfo"].(map[string]any)
	if info["name"] != "jaga-mcp" {
		t.Fatalf("serverInfo = %v", info)
	}
}

func TestInitializeDefaultVersion(t *testing.T) {
	s, _ := testServer(t)
	resp := s.handle(request("initialize", `1`, nil))
	result := resp.Result.(map[string]any)
	if result["protocolVersion"] != protocolVersion {
		t.Fatalf("protocolVersion = %v", result["protocolVersion"])
	}
}

func TestToolsList(t *testing.T) {
	s, _ := testServer(t)
	resp := s.handle(request("tools/list", `1`, nil))
	tools := resp.Result.(map[string]any)["tools"].([]toolDef)
	if len(tools) != 4 {
		t.Fatalf("jumlah tool = %d", len(tools))
	}
	want := map[string]bool{"jaga_check": false, "jaga_gate": false, "jaga_badge": false, "jaga_fix": false}
	for _, tool := range tools {
		if _, ok := want[tool.Name]; !ok {
			t.Fatalf("tool tak terduga: %s", tool.Name)
		}
		want[tool.Name] = true
		if tool.InputSchema["type"] != "object" {
			t.Fatalf("schema %s bukan object", tool.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("tool %s tidak ada", name)
		}
	}
}

func TestPingNotificationAndUnknownMethod(t *testing.T) {
	s, _ := testServer(t)
	if resp := s.handle(request("ping", `1`, nil)); resp.Error != nil {
		t.Fatal("ping gagal")
	}
	missingID := request("tools/list", `null`, nil)
	if resp := s.handle(missingID); resp != nil {
		t.Fatal("notifikasi tidak boleh dibalas")
	}
	resp := s.handle(request("nope", `1`, nil))
	if resp.Error == nil || resp.Error.Code != -32601 {
		t.Fatalf("error = %+v", resp.Error)
	}
}

func TestCheckFindings(t *testing.T) {
	s, dir := testServer(t)
	hygiene(t, dir)
	put(t, dir, ".env", "AWS_KEY=AKIA1234567890ABCDEF\n")

	text, isErr := callTool(t, s, "jaga_check", map[string]any{})
	if isErr {
		t.Fatalf("tool error: %s", text)
	}
	var out checkOutput
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		t.Fatal(err)
	}
	if out.Score >= 100 {
		t.Fatalf("skor = %d", out.Score)
	}
	if out.Summary["secret"] != 1 {
		t.Fatalf("summary = %v", out.Summary)
	}
	if out.Allowed {
		t.Fatal("seharusnya tidak allowed")
	}
	if len(out.Findings) != 1 || out.Findings[0].Check != "secret" {
		t.Fatalf("findings = %+v", out.Findings)
	}
	if strings.Contains(out.Findings[0].Message, "AKIA1234567890ABCDEF") {
		t.Fatal("secret tidak di-masking")
	}
}

func TestCheckFilter(t *testing.T) {
	s, dir := testServer(t)
	hygiene(t, dir)
	put(t, dir, "data.txt", "NIK: 3271010101010001\n")

	text, isErr := callTool(t, s, "jaga_check", map[string]any{"checks": []string{"pii"}})
	if isErr {
		t.Fatalf("tool error: %s", text)
	}
	var out checkOutput
	_ = json.Unmarshal([]byte(text), &out)
	if out.Summary["pii"] != 1 {
		t.Fatalf("summary = %v", out.Summary)
	}
	if _, ok := out.Summary["hygiene"]; ok {
		t.Fatalf("filter diabaikan: %v", out.Summary)
	}

	_, isErr = callTool(t, s, "jaga_check", map[string]any{"checks": []string{"nope"}})
	if !isErr {
		t.Fatal("check invalid seharusnya error")
	}
}

func TestGate(t *testing.T) {
	s, dir := testServer(t)
	hygiene(t, dir)
	put(t, dir, "data.txt", "NIK: 3271010101010001\n")

	text, isErr := callTool(t, s, "jaga_gate", map[string]any{"fail_on": "critical"})
	if isErr {
		t.Fatalf("tool error: %s", text)
	}
	var out gateOutput
	_ = json.Unmarshal([]byte(text), &out)
	if !out.Allowed || out.FailOn != "critical" {
		t.Fatalf("gate = %+v", out)
	}
	if len(out.Blocking) != 0 {
		t.Fatalf("blocking = %+v", out.Blocking)
	}

	text, isErr = callTool(t, s, "jaga_gate", map[string]any{"fail_on": "medium"})
	if isErr {
		t.Fatalf("tool error: %s", text)
	}
	_ = json.Unmarshal([]byte(text), &out)
	if out.Allowed || len(out.Blocking) == 0 {
		t.Fatalf("gate = %+v", out)
	}
}

func TestPathOutsideRoot(t *testing.T) {
	s, _ := testServer(t)
	text, isErr := callTool(t, s, "jaga_check", map[string]any{"path": "/etc"})
	if !isErr || !strings.Contains(text, "di luar root") {
		t.Fatalf("text=%q isErr=%v", text, isErr)
	}
}

func TestBadge(t *testing.T) {
	s, dir := testServer(t)
	hygiene(t, dir)
	text, isErr := callTool(t, s, "jaga_badge", map[string]any{})
	if isErr {
		t.Fatalf("tool error: %s", text)
	}
	if !strings.HasPrefix(text, "<svg") || !strings.Contains(text, "/100") {
		t.Fatalf("badge = %q", text)
	}
}

func TestServe(t *testing.T) {
	s, _ := testServer(t)
	in := strings.NewReader(strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"ping"}`,
	}, "\n") + "\n")
	var out bytes.Buffer
	if err := s.serve(in, &out); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("balasan = %d: %s", len(lines), out.String())
	}
	seen := map[string]bool{}
	for _, line := range lines {
		var resp rpcResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Fatal(err)
		}
		seen[string(resp.ID)] = true
	}
	if !seen["1"] || !seen["2"] {
		t.Fatalf("id balasan = %v", seen)
	}
}
