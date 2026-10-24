package markdown

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRenderBBGFMContract(t *testing.T) {
	source := "# Heading\n\n- [x] done\n- [ ] pending\n\n~~old~~ [safe](https://example.test/a?q=1)\n\n| A | B |\n|---|---|\n| 1 | 2 |"
	got, err := Render(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	wantHTML := "<h1>Heading</h1>\n<ul>\n<li><input checked=\"\" disabled=\"\" type=\"checkbox\"> done</li>\n<li><input disabled=\"\" type=\"checkbox\"> pending</li>\n</ul>\n<p><del>old</del> <a href=\"https://example.test/a?q=1\" rel=\"nofollow noreferrer\">safe</a></p>\n<table>\n<thead>\n<tr>\n<th>A</th>\n<th>B</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>1</td>\n<td>2</td>\n</tr>\n</tbody>\n</table>\n"
	if got.HTML != wantHTML {
		t.Fatalf("HTML mismatch\n got: %q\nwant: %q", got.HTML, wantHTML)
	}
	if got.VisibleText != "Heading done pending old safe A B 1 2" {
		t.Fatalf("visible text = %q", got.VisibleText)
	}
}

func TestRenderRejectsHostileAndOversizedOutput(t *testing.T) {
	got, err := Render(context.Background(), "safe <script>alert('private')</script> [bad](javascript:alert(1))")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"<script", "javascript:", "alert('private')"} {
		if strings.Contains(got.HTML, forbidden) {
			t.Fatalf("hostile output retained %q in %q", forbidden, got.HTML)
		}
	}
	for _, source := range []string{"", " \n", string([]byte{0xff}), strings.Repeat("x", MaximumSourceBytes+1)} {
		if _, err := Render(context.Background(), source); !errors.Is(err, ErrInvalidSource) {
			t.Fatalf("invalid source error = %v", err)
		}
	}
	if _, err := Render(context.Background(), strings.Repeat("&", MaximumSourceBytes)); !errors.Is(err, ErrOutputTooLarge) {
		t.Fatalf("expanding output error = %v", err)
	}
	if _, err := Render(context.Background(), "<script>only hostile content</script>"); !errors.Is(err, ErrNoRenderableContent) {
		t.Fatalf("empty sanitized output error = %v", err)
	}
}

func TestV1DoesNotEnableFutureProfiles(t *testing.T) {
	source := "$x^2$\n\n" + "```mermaid\ngraph TD; A-->B\n```"
	got, err := Render(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got.HTML, "<svg") || !strings.Contains(got.HTML, "<pre><code>graph TD; A--&gt;B") || !strings.Contains(got.VisibleText, "graph TD; A-->B") {
		t.Fatalf("future syntax was not preserved as ordinary Markdown: %+v", got)
	}
}

func TestRenderDeterministicAndCanceled(t *testing.T) {
	const source = "ordinary **content**"
	first, err := Render(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		next, err := Render(context.Background(), source)
		if err != nil || next != first {
			t.Fatalf("iteration %d: %+v %v", i, next, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Render(ctx, source); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled error = %v", err)
	}
}

func TestVisibleTextDoesNotDoubleDecodeEntities(t *testing.T) {
	got, err := Render(context.Background(), "literal &amp;lt; value")
	if err != nil {
		t.Fatal(err)
	}
	if got.VisibleText != "literal &lt; value" {
		t.Fatalf("visible text = %q", got.VisibleText)
	}
}

func TestRenderMatchesBBP2VisibleTextAndLinkPolicy(t *testing.T) {
	got, err := Render(context.Background(), "# Cafe\u0301\n\nnext\u00a0word\n\n[mail](mailto:private@example.test)\n\n---")
	if err != nil {
		t.Fatal(err)
	}
	if got.VisibleText != "Café next word mail" {
		t.Fatalf("visible text = %q", got.VisibleText)
	}
	if strings.Contains(got.HTML, "mailto:") {
		t.Fatalf("mailto link survived BB p2 policy: %q", got.HTML)
	}
	empty, err := Render(context.Background(), "---")
	if err != nil || empty.HTML != "<hr>\n" || empty.VisibleText != "" {
		t.Fatalf("semantic empty-search document = %+v, %v", empty, err)
	}
}

func TestRendererIdentityBindsVersionAndConfiguration(t *testing.T) {
	configDigest := strings.Repeat("a", 64)
	a := RendererIdentity("1.0.0", configDigest)
	b := RendererIdentity("1.0.0", strings.Repeat("b", 64))
	prerelease := RendererIdentity("1.0.0-beta.1", configDigest)
	const wantJSON = `{"schema":"gotth.markdown.render-response.v1","extension_version":"1.0.0","engine":"goldmark-v1.8.5-gfm","profile":"gotth.markdown.bb-gfm.v1","sanitizer":"bluemonday-v1.0.27-bb-p2","configuration_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","assets":"none","maximum_source_bytes":65536,"maximum_html_bytes":262144,"maximum_visible_text_bytes":65536}`
	if got := string(RendererIdentityCanonicalJSON("1.0.0", configDigest)); got != wantJSON {
		t.Fatalf("identity JSON = %q", got)
	}
	if a != "f5050a806c1f941b8a806ae816fd1fdf358267d60b745909cde54c5ec7499820" || a == b || a == prerelease || a != RendererIdentity("1.0.0", configDigest) {
		t.Fatalf("identity mismatch: %q %q %q", a, b, prerelease)
	}
}

func FuzzRenderNeverReturnsPartialSuccess(f *testing.F) {
	for _, seed := range []string{"hello", "<script>x</script>", "[x](javascript:x)", strings.Repeat("&", 100)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, source string) {
		if len(source) > MaximumSourceBytes {
			return
		}
		got, err := Render(context.Background(), source)
		if err == nil && (got.HTML == "" || len(got.HTML) > MaximumHTMLBytes || len(got.VisibleText) > MaximumVisibleTextBytes) {
			t.Fatalf("partial success: %+v", got)
		}
	})
}

func BenchmarkRenderRepresentative(b *testing.B) {
	source := strings.Repeat("## Heading\n\nA [link](https://example.test) and **strong** text.\n\n", 500)
	benchmarkRender(b, source)
}

func BenchmarkRenderMaximumSource(b *testing.B) {
	source := strings.Repeat("a ", MaximumSourceBytes/2)
	benchmarkRender(b, source)
}

func benchmarkRender(b *testing.B, source string) {
	b.Helper()
	b.ReportAllocs()
	b.SetBytes(int64(len(source)))
	for i := 0; i < b.N; i++ {
		if _, err := Render(context.Background(), source); err != nil {
			b.Fatal(err)
		}
	}
}
