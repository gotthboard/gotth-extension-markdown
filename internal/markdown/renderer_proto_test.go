package markdown

import (
	"testing"

	rendererv1 "github.com/gotthboard/gotth-extension-markdown/proto/gotth/markdown/renderer/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestRendererWireDescriptorIsNarrowAndVersioned(t *testing.T) {
	file := rendererv1.File_proto_gotth_markdown_renderer_v1_renderer_proto
	if file.Path() != "proto/gotth/markdown/renderer/v1/renderer.proto" || string(file.Package()) != "gotth.markdown.renderer.v1" {
		t.Fatalf("renderer descriptor identity = %q %q", file.Path(), file.Package())
	}
	services := file.Services()
	if services.Len() != 1 || string(services.Get(0).FullName()) != "gotth.markdown.renderer.v1.MarkdownRenderer" || services.Get(0).Methods().Len() != 1 {
		t.Fatalf("unexpected renderer service: %+v", services)
	}
	if rendererv1.MarkdownRenderer_Render_FullMethodName != "/gotth.markdown.renderer.v1.MarkdownRenderer/Render" {
		t.Fatal("renderer RPC path drifted")
	}
	messages := file.Messages()
	if messages.Len() != 2 {
		t.Fatalf("message count = %d", messages.Len())
	}
	want := map[protoreflect.FullName]map[protoreflect.FieldNumber]protoreflect.Kind{
		"gotth.markdown.renderer.v1.RenderRequest":  {1: protoreflect.StringKind, 2: protoreflect.StringKind, 3: protoreflect.StringKind},
		"gotth.markdown.renderer.v1.RenderResponse": {1: protoreflect.StringKind, 2: protoreflect.BoolKind, 3: protoreflect.StringKind, 4: protoreflect.StringKind, 5: protoreflect.StringKind, 6: protoreflect.StringKind, 7: protoreflect.StringKind, 8: protoreflect.StringKind, 9: protoreflect.StringKind, 10: protoreflect.StringKind},
	}
	for i := 0; i < messages.Len(); i++ {
		message := messages.Get(i)
		expected, ok := want[message.FullName()]
		if !ok || message.Fields().Len() != len(expected) {
			t.Fatalf("unexpected message %s", message.FullName())
		}
		for j := 0; j < message.Fields().Len(); j++ {
			field := message.Fields().Get(j)
			if expected[field.Number()] != field.Kind() || field.IsMap() || field.Cardinality() == protoreflect.Repeated {
				t.Fatalf("unexpected field %s number=%d kind=%s", field.FullName(), field.Number(), field.Kind())
			}
		}
	}
}
