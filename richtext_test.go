package richtext

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitizeStripsXSS(t *testing.T) {
	vectors := []string{
		`<script>alert(1)</script>`,
		`<img src=x onerror="alert(1)">`,
		`<a href="javascript:alert(1)">click</a>`,
		`<div onclick="evil()">x</div>`,
		`<iframe src="evil"></iframe>`,
	}
	for _, v := range vectors {
		out := Sanitize(v)
		low := strings.ToLower(out)
		if strings.Contains(low, "<script") || strings.Contains(low, "onerror") ||
			strings.Contains(low, "onclick") || strings.Contains(low, "javascript:") ||
			strings.Contains(low, "<iframe") {
			t.Errorf("XSS not neutralized: %q → %q", v, out)
		}
	}
}

func TestSanitizeKeepsSafeTags(t *testing.T) {
	out := Sanitize(`<p>Hello <strong>world</strong> <a href="https://example.com">link</a></p>`)
	for _, want := range []string{"<p>", "<strong>world</strong>", `href="https://example.com"`} {
		if !strings.Contains(out, want) {
			t.Errorf("safe content dropped: missing %q in %q", want, out)
		}
	}
}

func TestRenderMarkdown(t *testing.T) {
	html, err := RenderMarkdown("# Title\n\nSome **bold** and a [link](https://x.com).")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "<h1>") || !strings.Contains(html, "<strong>bold</strong>") {
		t.Fatalf("markdown not rendered: %q", html)
	}
	// Embedded HTML in markdown must still be sanitized.
	bad, _ := RenderMarkdown("Hi\n\n<script>alert(1)</script>")
	if strings.Contains(strings.ToLower(bad), "<script") {
		t.Fatalf("script survived markdown render: %q", bad)
	}
}

func TestPlainTextAndExcerpt(t *testing.T) {
	rt := New(`<h1>Hello</h1><p>This is <strong>rich</strong> &amp; safe content for testing excerpts.</p>`)
	pt := rt.PlainText()
	if strings.Contains(pt, "<") {
		t.Fatalf("PlainText left tags: %q", pt)
	}
	if !strings.Contains(pt, "Hello") || !strings.Contains(pt, "&") {
		t.Fatalf("PlainText wrong: %q", pt)
	}
	ex := rt.Excerpt(12)
	if !strings.HasSuffix(ex, "…") || len([]rune(ex)) > 16 {
		t.Fatalf("Excerpt wrong: %q", ex)
	}
	if New("").IsEmpty() != true {
		t.Error("empty content should be IsEmpty")
	}
}

func TestFromMarkdownAndJSON(t *testing.T) {
	rt := FromMarkdown("# Hi\n\n**bold**")
	if rt.Raw != "# Hi\n\n**bold**" || !strings.Contains(rt.HTML, "<h1>Hi</h1>") {
		t.Fatalf("FromMarkdown = %+v", rt)
	}
	b, err := json.Marshal(rt)
	if err != nil {
		t.Fatal(err)
	}
	var back RichText
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.Raw != rt.Raw || back.HTML != rt.HTML {
		t.Fatalf("JSON round-trip: %+v vs %+v", back, rt)
	}
	// Unmarshal must re-sanitize a tampered html field.
	var tampered RichText
	if err := json.Unmarshal([]byte(`{"raw":"x","html":"<script>alert(1)</script><p>ok</p>"}`), &tampered); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(tampered.HTML), "<script") {
		t.Fatalf("unmarshal did not re-sanitize: %q", tampered.HTML)
	}
	// bare string unmarshals as HTML
	var s RichText
	if err := json.Unmarshal([]byte(`"<b>hi</b>"`), &s); err != nil || !strings.Contains(s.HTML, "<b>hi</b>") {
		t.Fatalf("string unmarshal = %+v, %v", s, err)
	}
}
