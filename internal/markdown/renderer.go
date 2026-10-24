package markdown

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"golang.org/x/net/html"
	"golang.org/x/text/unicode/norm"
)

var (
	ErrInvalidSource       = errors.New("invalid Markdown source")
	ErrNoRenderableContent = errors.New("Markdown has no renderable content")
	ErrOutputTooLarge      = errors.New("rendered output exceeds limit")
	renderer               = goldmark.New(goldmark.WithExtensions(
		extension.NewLinkify(extension.WithLinkifyAllowedProtocols([]string{"http:", "https:"})),
		extension.NewTable(extension.WithTableCellAlignMethod(extension.TableCellAlignNone)),
		extension.Strikethrough,
		extension.TaskList,
	))
	sanitizer = newSanitizer()
)

type Result struct {
	HTML        string
	VisibleText string
}

type RendererIdentityDocument struct {
	Schema                  string `json:"schema"`
	ExtensionVersion        string `json:"extension_version"`
	Engine                  string `json:"engine"`
	Profile                 string `json:"profile"`
	Sanitizer               string `json:"sanitizer"`
	ConfigurationSHA256     string `json:"configuration_sha256"`
	Assets                  string `json:"assets"`
	MaximumSourceBytes      int    `json:"maximum_source_bytes"`
	MaximumHTMLBytes        int    `json:"maximum_html_bytes"`
	MaximumVisibleTextBytes int    `json:"maximum_visible_text_bytes"`
}

func RendererIdentityCanonicalJSON(extensionVersion, configurationSHA256 string) []byte {
	encoded, err := json.Marshal(RendererIdentityDocument{ResponseSchema, extensionVersion, EngineIdentity, ProfileBBGFM, SanitizerIdentity, configurationSHA256, "none", MaximumSourceBytes, MaximumHTMLBytes, MaximumVisibleTextBytes})
	if err != nil {
		panic(err)
	}
	return encoded
}

func RendererIdentity(extensionVersion, configurationSHA256 string) string {
	digest := sha256.Sum256(RendererIdentityCanonicalJSON(extensionVersion, configurationSHA256))
	return hex.EncodeToString(digest[:])
}

func Render(ctx context.Context, source string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if len(source) == 0 || len(source) > MaximumSourceBytes || !utf8.ValidString(source) || strings.TrimSpace(source) == "" {
		return Result{}, ErrInvalidSource
	}
	var raw bytes.Buffer
	if err := renderer.Convert([]byte(source), &raw); err != nil {
		return Result{}, fmt.Errorf("render Markdown: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	htmlOutput := sanitizer.Sanitize(filterTaskListInputs(raw.String()))
	if len(htmlOutput) == 0 || strings.TrimSpace(htmlOutput) == "" {
		return Result{}, ErrNoRenderableContent
	}
	if len(htmlOutput) > MaximumHTMLBytes {
		return Result{}, ErrOutputTooLarge
	}
	visible, err := projectVisibleText(htmlOutput)
	if err != nil {
		return Result{}, fmt.Errorf("project visible text: %w", err)
	}
	if len(visible) > MaximumVisibleTextBytes {
		return Result{}, ErrOutputTooLarge
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	return Result{HTML: htmlOutput, VisibleText: visible}, nil
}

func newSanitizer() *bluemonday.Policy {
	policy := bluemonday.NewPolicy()
	policy.AllowElements("p", "h1", "h2", "h3", "h4", "h5", "h6", "hr", "em", "strong", "ul", "ol", "li", "a", "blockquote", "pre", "code", "br", "table", "thead", "tbody", "tr", "th", "td", "del", "input")
	policy.AllowAttrs("type").Matching(regexp.MustCompile(`^checkbox$`)).OnElements("input")
	policy.AllowAttrs("disabled", "checked").Matching(regexp.MustCompile(`^$`)).OnElements("input")
	policy.AllowAttrs("href").OnElements("a")
	policy.AllowRelativeURLs(true)
	policy.AllowURLSchemes("http", "https")
	policy.RequireNoFollowOnLinks(true)
	policy.RequireNoReferrerOnLinks(true)
	return policy
}

func filterTaskListInputs(raw string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(raw))
	var output bytes.Buffer
	output.Grow(len(raw))
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			return output.String()
		}
		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			rawToken := tokenizer.Raw()
			if isRawInputTag(rawToken) {
				rawInput := append([]byte(nil), rawToken...)
				token := tokenizer.Token()
				if !containsASCIIUpper(rawInput) && validTaskListInput(token.Attr) {
					output.Write(rawInput)
				}
				continue
			}
		}
		output.Write(tokenizer.Raw())
	}
}
func isRawInputTag(raw []byte) bool {
	const n = len("input")
	if len(raw) < 1+n || raw[0] != '<' || !bytes.EqualFold(raw[1:1+n], []byte("input")) {
		return false
	}
	if len(raw) == 1+n {
		return true
	}
	switch raw[1+n] {
	case ' ', '\t', '\n', '\r', '\f', '/', '>':
		return true
	}
	return false
}
func containsASCIIUpper(value []byte) bool {
	for _, c := range value {
		if c >= 'A' && c <= 'Z' {
			return true
		}
	}
	return false
}
func validTaskListInput(attributes []html.Attribute) bool {
	typeSeen, disabledSeen, checkedSeen := false, false, false
	for _, attribute := range attributes {
		switch attribute.Key {
		case "type":
			if typeSeen || attribute.Val != "checkbox" {
				return false
			}
			typeSeen = true
		case "disabled":
			if disabledSeen || attribute.Val != "" {
				return false
			}
			disabledSeen = true
		case "checked":
			if checkedSeen || attribute.Val != "" {
				return false
			}
			checkedSeen = true
		default:
			return false
		}
	}
	return typeSeen && disabledSeen
}

func projectVisibleText(rendered string) (string, error) {
	tokenizer := html.NewTokenizer(strings.NewReader(rendered))
	var output strings.Builder
	output.Grow(len(rendered))
	wroteText := false
	pendingSpace := false
	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			if err := tokenizer.Err(); err != nil && !errors.Is(err, io.EOF) {
				return "", err
			}
			return norm.NFC.String(output.String()), nil
		case html.TextToken:
			for _, character := range tokenizer.Token().Data {
				if isUnicode15WhiteSpace(character) {
					pendingSpace = wroteText
					continue
				}
				if pendingSpace {
					output.WriteByte(' ')
					pendingSpace = false
				}
				output.WriteRune(character)
				wroteText = true
			}
		case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
			name, _ := tokenizer.TagName()
			if isVisibleTextBoundary(string(name)) {
				pendingSpace = wroteText
			}
		}
	}
}

func isVisibleTextBoundary(name string) bool {
	switch name {
	case "p", "h1", "h2", "h3", "h4", "h5", "h6", "hr", "ul", "ol", "li", "blockquote", "pre", "br", "table", "thead", "tbody", "tr", "th", "td":
		return true
	default:
		return false
	}
}

func isUnicode15WhiteSpace(character rune) bool {
	switch character {
	case '\u0009', '\u000a', '\u000b', '\u000c', '\u000d', '\u0020', '\u0085', '\u00a0', '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000':
		return true
	default:
		return character >= '\u2000' && character <= '\u200a'
	}
}
