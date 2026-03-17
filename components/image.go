package components

func init() {
	Register(ComponentDef{
		Tag: "Image",
		ExtractFrame: func(fd *FrameData, node NodeInfo, children []NodeInfo) {
			passthrough := []string{
				"src", "contentMode", "placeholder", "fadeDuration",
				"showActivityIndicator", "tintColor",
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
