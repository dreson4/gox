package components

import "gox/internal/yoga"

func init() {
	Register(ComponentDef{
		Tag: "Switch",
		ConfigureYoga: func(yn *yoga.Node, props map[string]any, children []NodeInfo, screen ScreenInfoData) {
			yn.SetWidth(51)  // iOS UISwitch default
			yn.SetHeight(31)
		},
	})
}
