package check

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Emil-02/jaga/internal/config"
)

func defaultCfg() config.Config { return config.Default() }

func TestSecretRun(t *testing.T) {
	dir := t.TempDir()
	awsKey := "AKIA" + strings.Repeat("A1B2C3D4", 2)
	content := "aws := \"" + awsKey + "\"\nclean := \"hello\"\n"
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	res := Secret{}.Run(context.Background(), dir, defaultCfg())
	if len(res.Findings) == 0 {
		t.Fatal("secret tidak terdeteksi")
	}
	if strings.Contains(res.Findings[0].Message, awsKey) {
		t.Fatal("secret tidak di-mask")
	}
}

func TestSecretVibeCodingAndAI(t *testing.T) {
	dir := t.TempDir()

	googleKey := "AIzaSy" + strings.Repeat("B", 32) + "Q"
	anthropicKey := "sk-ant-api03-" + strings.Repeat("a", 40)
	supabaseKey := "sb_secret_" + strings.Repeat("1", 20)
	jwtHeader := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	anonPayload := base64.RawURLEncoding.EncodeToString([]byte(`{"role":"anon","iss":"supabase"}`))
	servicePayload := base64.RawURLEncoding.EncodeToString([]byte(`{"role":"service_role","iss":"supabase"}`))
	jwtSig := strings.Repeat("s", 20)
	anonToken := jwtHeader + "." + anonPayload + "." + jwtSig
	serviceToken := jwtHeader + "." + servicePayload + "." + jwtSig

	files := map[string]string{
		".env.local": strings.Join([]string{
			"GEMINI_KEY=" + googleKey,
			"ANTHROPIC_KEY=" + anthropicKey,
			"MIDTRANS_SERVER_KEY=SB-Mid-server-" + strings.Repeat("m", 24),
			"DATABASE_URL=postgres://postgres:" + strings.Repeat("p", 12) + "@db.supabase.co:5432/postgres",
			"NEXT_PUBLIC_SERVICE_ROLE_KEY=" + supabaseKey,
			"PLACEHOLDER_KEY=your_api_key_here",
		}, "\n"),
		"client.ts": strings.Join([]string{
			"const anonToken = \"" + anonToken + "\";",
			"const leakedServiceRole = \"" + serviceToken + "\";",
		}, "\n"),
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	res := Secret{}.Run(context.Background(), dir, defaultCfg())
	found := map[string]bool{}
	for _, f := range res.Findings {
		found[f.Message] = true
	}

	expectedPrefixes := []string{
		"Google Gemini/Cloud API key:",
		"Anthropic API key:",
		"Midtrans Server key:",
		"Database URI with password:",
		"Vibe-code frontend secret leak:",
		"Supabase service_role key:",
	}

	for _, prefix := range expectedPrefixes {
		matched := false
		for msg := range found {
			if strings.HasPrefix(msg, prefix) {
				matched = true
				break
			}
		}
		if !matched {
			t.Fatalf("ekspektasi deteksi %q tetapi tidak ditemukan. Temuan: %+v", prefix, res.Findings)
		}
	}

	// Pastikan placeholder tidak memicu finding
	for msg := range found {
		if strings.Contains(msg, "your_api_key_here") {
			t.Fatalf("placeholder tidak seharusnya dideteksi: %s", msg)
		}
	}
}

func TestPIIRun(t *testing.T) {
	dir := t.TempDir()
	content := "nik: 3204123456789012\nhp: 081234567890\n"
	if err := os.WriteFile(filepath.Join(dir, "data.txt"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	res := PII{}.Run(context.Background(), dir, defaultCfg())
	if len(res.Findings) < 2 {
		t.Fatalf("hanya %d temuan: %+v", len(res.Findings), res.Findings)
	}
	for _, f := range res.Findings {
		if strings.Contains(f.Message, "3204123456789012") || strings.Contains(f.Message, "081234567890") {
			t.Fatal("data pribadi tidak di-mask")
		}
	}
}

func TestHygieneRun(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}
	res := Hygiene{}.Run(context.Background(), dir, defaultCfg())
	if len(res.Findings) != 3 {
		t.Fatalf("harusnya 3 kekurangan, dapat %d: %+v", len(res.Findings), res.Findings)
	}
}
