package components

import "github.com/dreson4/gox/internal/yoga"

func init() {
	Register(ComponentDef{
		Tag: "TextInput",
		ConfigureYoga: func(yn *yoga.Node, props map[string]any, children []NodeInfo, screen ScreenInfoData) {
			// Default iOS text field height if no explicit style height
			if s, ok := props["style"]; ok {
				if style, ok := s.(interface{ GetHeight() float64 }); ok {
					if style.GetHeight() > 0 {
						return // explicit height set, don't override
					}
				}
			}
			yn.SetHeight(44)
		},
		ExtractFrame: func(fd *FrameData, node NodeInfo, children []NodeInfo) {
			// Pass through TextInput-specific props to the layout frame
			passthrough := []string{
				"value", "placeholder", "placeholderColor",
				"keyboardType", "returnKeyType", "autoCapitalize",
				"autoCorrect", "autoFocus", "editable", "maxLength",
				"secure", "textAlign", "color", "fontSize",
			}
			for _, key := range passthrough {
				if v, ok := node.Props[key]; ok {
					if fd.Props == nil {
						fd.Props = map[string]any{}
					}
					fd.Props[key] = v
				}
			}
		},
	})
}
