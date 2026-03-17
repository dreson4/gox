package components

func init() {
	Register(ComponentDef{
		Tag: "Button",
		ExtractFrame: func(frame *FrameData, node NodeInfo, children []NodeInfo) {
			for _, child := range children {
				if child.Type == NodeTypeElement && child.Tag == "Text" {
					frame.Text = collectText(child.Children)
					// Copy child text styling for bridge to apply to UIButton title
					if childStyle, ok := child.Props["style"]; ok {
						if frame.Props == nil {
							frame.Props = map[string]any{}
						}
						if s, ok := childStyle.(interface {
							GetColor() string
							GetFontSize() float64
							GetFontWeight() string
						}); ok {
							frame.Props["_btnTextColor"] = s.GetColor()
							frame.Props["_btnTextSize"] = s.GetFontSize()
							frame.Props["_btnTextWeight"] = s.GetFontWeight()
						}
					}
					break
				}
			}
		},
	})
}
