package components

import "gox/internal/yoga"

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
	})
}
