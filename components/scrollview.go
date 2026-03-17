package components

import "gox/internal/yoga"

func init() {
	Register(ComponentDef{
		Tag: "ScrollView",
		ConfigureYoga: func(yn *yoga.Node, props map[string]any, children []NodeInfo, screen ScreenInfoData) {
			// ScrollView uses overflow:scroll so Yoga allows children to exceed parent bounds
			yn.SetOverflow(yoga.OverflowScroll)

			// Horizontal scroll mode: layout children in a row
			if horizontal, ok := props["horizontal"].(bool); ok && horizontal {
				yn.SetFlexDirection(yoga.FlexDirectionRow)
			}
		},
		ExtractFrame: func(fd *FrameData, node NodeInfo, children []NodeInfo) {
			passthrough := []string{
				"horizontal", "scrollEnabled", "bounces",
				"alwaysBounceVertical", "alwaysBounceHorizontal",
				"pagingEnabled", "showsVerticalIndicator", "showsHorizontalIndicator",
				"decelerationRate", "keyboardDismissMode",
				"contentInsetTop", "contentInsetBottom",
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
