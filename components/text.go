package components

import "github.com/dreson4/gox/internal/yoga"

func init() {
	Register(ComponentDef{
		Tag:          "Text",
		SkipChildren: true,
		ConfigureYoga: func(yn *yoga.Node, props map[string]any, children []NodeInfo, screen ScreenInfoData) {
			text := collectText(children)
			fontSize := 17.0
			if s, ok := props["style"]; ok {
				if style, ok := s.(interface{ GetFontSize() float64 }); ok {
					if fs := style.GetFontSize(); fs > 0 {
						fontSize = fs
					}
				}
			}
			EstimateTextSize(yn, text, fontSize)
		},
		ExtractFrame: func(frame *FrameData, node NodeInfo, children []NodeInfo) {
			frame.Text = collectText(children)
		},
	})
}

func collectText(children []NodeInfo) string {
	var text string
	for _, child := range children {
		if child.Type == NodeTypeText {
			text += child.Text
		}
	}
	return text
}

// EstimateTextSize sets rough height on a Yoga node based on text content.
// Exported so layout.go can also use it for raw text nodes.
func EstimateTextSize(yn *yoga.Node, text string, fontSize float64) {
	if text == "" {
		return
	}
	lineHeight := fontSize * 1.4
	yn.SetHeight(float32(lineHeight))
}
