package components

import "github.com/dreson4/gox/internal/yoga"

func init() {
	Register(ComponentDef{
		Tag: "SafeArea",
		ConfigureYoga: func(yn *yoga.Node, props map[string]any, children []NodeInfo, screen ScreenInfoData) {
			yn.SetPadding(yoga.EdgeTop, float32(screen.SafeTop))
			yn.SetPadding(yoga.EdgeRight, float32(screen.SafeRight))
			yn.SetPadding(yoga.EdgeBottom, float32(screen.SafeBottom))
			yn.SetPadding(yoga.EdgeLeft, float32(screen.SafeLeft))
		},
	})
}
