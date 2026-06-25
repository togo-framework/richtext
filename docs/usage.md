# richtext — usage

## Build a value
```go
rt := richtext.New(untrustedHTML)        // sanitized HTML in rt.HTML
rt := richtext.FromMarkdown(markdown)    // md → sanitized HTML
```

## Derive
```go
rt.PlainText()   // tags stripped, entities decoded, whitespace collapsed
rt.Excerpt(160)  // preview, word-boundary truncation + ellipsis
rt.IsEmpty()
```

## One-off helpers
```go
richtext.Sanitize(html)        // clean untrusted HTML
richtext.RenderMarkdown(md)    // (string, error) — md → safe HTML
richtext.StripTags(html)       // → plain text
```

## JSON
`RichText` is `{"raw","html"}`; **unmarshal re-sanitizes** the html field (never trust stored html). A bare JSON string is treated as HTML.

## REST
`POST /api/richtext/render` `{"markdown"|"html": "..."}` → `{"html","text"}`.

## Security
Allowlist via bluemonday UGC policy; Markdown via goldmark then sanitized. `<script>`, event handlers (`onerror`/`onclick`), `javascript:` URLs and `<iframe>` are removed. Store + render the sanitized HTML, not raw input.
