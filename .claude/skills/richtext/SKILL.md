---
name: richtext
description: Handle rich-text/Markdown content safely in a togo app with the richtext plugin — sanitize untrusted HTML (XSS-safe), render Markdown to HTML, and extract plain-text excerpts.
---

# togo richtext

Use this to accept and render user rich text without XSS.

## Store
```go
rt := richtext.New(editorHTML)        // sanitized → rt.HTML (safe to render)
rt := richtext.FromMarkdown(md)       // md → sanitized HTML
```
Persist the `RichText` ({raw, html}); it **re-sanitizes on unmarshal**.

## Use
- Render `rt.HTML` (already safe). Never render `rt.Raw`.
- `rt.PlainText()` for search indexing; `rt.Excerpt(n)` for list previews.
- `POST /api/richtext/render` to render md/html → safe html for a client editor.

## Rules
- Always sanitize on the way IN and trust only `rt.HTML` on the way OUT.
- bluemonday UGC policy strips scripts, event handlers, javascript: URLs, iframes.
- For Markdown, use `FromMarkdown`/`RenderMarkdown` (renders then sanitizes).
