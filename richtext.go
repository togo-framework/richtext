// Package richtext handles rich-text content for togo (Action Text / Trix
// style): sanitize untrusted HTML to an XSS-safe allowlist, render Markdown to
// safe HTML, and derive plain-text excerpts for search/previews.
//
//	rt := richtext.FromMarkdown("# Hi\n\n<script>alert(1)</script> **bold**")
//	rt.HTML        // sanitized: "<h1>Hi</h1>\n<p> <strong>bold</strong></p>"
//	rt.PlainText() // "Hi  bold"
package richtext

import (
	"bytes"
	"encoding/json"
	"html"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/microcosm-cc/bluemonday"
	"github.com/togo-framework/togo"
	"github.com/yuin/goldmark"
)

// policy is the HTML allowlist applied to all stored content (user-generated
// content policy: common formatting tags, safe links/images — no scripts,
// event handlers, or javascript: URLs).
var policy = bluemonday.UGCPolicy()

// strict strips every tag, leaving text only (for plain-text extraction).
var strict = bluemonday.StrictPolicy()

// mark renders Markdown → HTML (CommonMark).
var mark = goldmark.New()

// Sanitize cleans untrusted HTML to the safe allowlist.
func Sanitize(unsafeHTML string) string { return policy.Sanitize(unsafeHTML) }

// RenderMarkdown converts Markdown to sanitized HTML.
func RenderMarkdown(md string) (string, error) {
	var buf bytes.Buffer
	if err := mark.Convert([]byte(md), &buf); err != nil {
		return "", err
	}
	return policy.Sanitize(buf.String()), nil
}

// StripTags returns the plain text of an HTML fragment (tags removed, entities
// decoded, whitespace collapsed).
func StripTags(htmlStr string) string {
	text := html.UnescapeString(strict.Sanitize(htmlStr))
	return strings.Join(strings.Fields(text), " ")
}

// RichText is a stored rich-text value: the raw input plus the sanitized HTML
// that is safe to render.
type RichText struct {
	Raw  string `json:"raw"`
	HTML string `json:"html"`
}

// New builds RichText from (untrusted) HTML, storing the sanitized result.
func New(rawHTML string) RichText {
	return RichText{Raw: rawHTML, HTML: Sanitize(rawHTML)}
}

// FromMarkdown builds RichText from Markdown source.
func FromMarkdown(md string) RichText {
	out, err := RenderMarkdown(md)
	if err != nil {
		out = ""
	}
	return RichText{Raw: md, HTML: out}
}

// PlainText returns the rendered HTML stripped to plain text.
func (r RichText) PlainText() string { return StripTags(r.HTML) }

// Excerpt returns the first n characters of the plain text, with an ellipsis
// when truncated (cuts on a word boundary where possible).
func (r RichText) Excerpt(n int) string {
	t := r.PlainText()
	if n <= 0 || len(t) <= n {
		return t
	}
	cut := t[:n]
	if i := strings.LastIndex(cut, " "); i > n/2 {
		cut = cut[:i]
	}
	return strings.TrimSpace(cut) + "…"
}

// IsEmpty reports whether there is no visible content.
func (r RichText) IsEmpty() bool { return strings.TrimSpace(r.PlainText()) == "" }

// MarshalJSON / UnmarshalJSON keep raw + (re-sanitized) html in sync.
func (r RichText) MarshalJSON() ([]byte, error) {
	type alias RichText
	return json.Marshal(alias(r))
}

func (r *RichText) UnmarshalJSON(b []byte) error {
	type alias RichText
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		// allow a bare string (treated as HTML)
		var s string
		if err2 := json.Unmarshal(b, &s); err2 != nil {
			return err
		}
		*r = New(s)
		return nil
	}
	// Always re-sanitize on the way in — never trust a stored html field.
	r.Raw = a.Raw
	r.HTML = Sanitize(a.HTML)
	if r.HTML == "" && a.Raw != "" {
		r.HTML = Sanitize(a.Raw)
	}
	return nil
}

// --- optional kernel plugin: a render endpoint ---

// Service is the (stateless) richtext service stored on the kernel.
type Service struct{}

func init() {
	togo.RegisterProviderFunc("richtext", togo.PriorityService, func(k *togo.Kernel) error {
		s := &Service{}
		k.Set("richtext", s)
		if k.Router != nil {
			s.mountRoutes(k.Router)
		}
		return nil
	})
}

// FromKernel returns the richtext Service.
func FromKernel(k *togo.Kernel) (*Service, bool) {
	v, ok := k.Get("richtext")
	if !ok {
		return nil, false
	}
	s, ok := v.(*Service)
	return s, ok
}

func (s *Service) mountRoutes(r chi.Router) {
	// POST /api/richtext/render { "markdown": "..." } or { "html": "..." } → safe HTML.
	r.Post("/api/richtext/render", func(w http.ResponseWriter, req *http.Request) {
		var in struct {
			Markdown string `json:"markdown"`
			HTML     string `json:"html"`
		}
		_ = json.NewDecoder(req.Body).Decode(&in)
		var rt RichText
		if in.Markdown != "" {
			rt = FromMarkdown(in.Markdown)
		} else {
			rt = New(in.HTML)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"html": rt.HTML, "text": rt.PlainText()})
	})
}
