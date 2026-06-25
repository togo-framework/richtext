<div align="center">
  <img src=".github/assets/togo-mark.svg" alt="togo" height="64" />
  <h1>togo-framework/richtext</h1>
  <p>
    <a href="https://to-go.dev/marketplace"><img src="https://img.shields.io/badge/marketplace-to--go.dev-1FC7DC" alt="marketplace" /></a>
    <a href="https://pkg.go.dev/github.com/togo-framework/richtext"><img src="https://pkg.go.dev/badge/github.com/togo-framework/richtext.svg" alt="pkg.go.dev" /></a>
    <img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT" />
  </p>
  <p><strong>Rich-text for <a href="https://to-go.dev">togo</a> — sanitize, render Markdown, store XSS-safe HTML.</strong></p>
</div>

## Install

```bash
togo install togo-framework/richtext
```

The togo answer to Rails **Action Text** / Trix. Accept user rich text, **sanitize** it to a safe allowlist (no `<script>`, event handlers, or `javascript:` URLs), render **Markdown → safe HTML**, and derive plain-text excerpts for search/previews.

## Usage

```go
// From untrusted HTML (e.g. a WYSIWYG editor):
rt := richtext.New(userHTML)        // rt.HTML is sanitized + safe to render
// From Markdown:
rt = richtext.FromMarkdown("# Hi\n\n**bold** <script>alert(1)</script>")
rt.HTML        // "<h1>Hi</h1>\n<p><strong>bold</strong> </p>"  (script removed)
rt.PlainText() // "Hi bold"  — tags stripped, entities decoded
rt.Excerpt(120)// preview text, ellipsised on a word boundary

richtext.Sanitize(htmlStr)         // one-off clean
richtext.RenderMarkdown(md)        // md → sanitized html
```

`RichText` marshals as `{"raw": "...", "html": "..."}` and **re-sanitizes on unmarshal** — a stored `html` field is never trusted.

## REST

`POST /api/richtext/render` with `{"markdown":"..."}` or `{"html":"..."}` → `{"html": "<safe>", "text": "..."}`.

## Security

HTML is sanitized with [bluemonday](https://github.com/microcosm-cc/bluemonday)'s UGC policy; Markdown is rendered with [goldmark](https://github.com/yuin/goldmark) and then sanitized. Always store and re-render the **sanitized** HTML; never render raw user input.

---

<div align="center">
  <h3>Premium sponsors</h3>
  <p>
    <a href="https://id8media.com"><strong>ID8 Media</strong></a> &nbsp;·&nbsp;
    <a href="https://one-studio.co"><strong>One Studio</strong></a>
  </p>
  <p><sub>Support togo — <a href="https://github.com/sponsors/fadymondy">become a sponsor</a>.</sub></p>
</div>
