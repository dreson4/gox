package components

import "github.com/dreson4/gox/internal/yoga"

// ComponentDef defines a component's layout behavior on the Go side.
// Each built-in and third-party component registers one of these.
type ComponentDef struct {
	Tag string

	// SkipChildren: if true, buildYogaTree won't recurse into children.
	// Used by Text (which estimates size from text content instead).
	SkipChildren bool

	// ConfigureYoga sets intrinsic sizes or special padding on the Yoga node.
	// Called during buildYogaTree after style is applied.
	ConfigureYoga func(yn *yoga.Node, props map[string]any, children []NodeInfo, screen ScreenInfoData)

	// ExtractFrame customizes the LayoutFrame output.
	// Called during extractFrames for this tag.
	ExtractFrame func(frame *FrameData, node NodeInfo, children []NodeInfo)
}

// ScreenInfoData is passed to ConfigureYoga so components can read screen dimensions.
type ScreenInfoData struct {
	Width, Height                          float64
	SafeTop, SafeRight, SafeBottom, SafeLeft float64
}

// NodeInfo provides read-only access to a node for component callbacks.
type NodeInfo struct {
	Tag      string
	Text     string
	Type     int // 0=Element, 1=Text, 2=Fragment
	Props    map[string]any
	Children []NodeInfo
}

// FrameData is the mutable frame passed to ExtractFrame callbacks.
type FrameData struct {
	Text  string
	Props map[string]any
}

// Node type constants matching gox.NodeType values.
const (
	NodeTypeElement  = 0
	NodeTypeText     = 1
	NodeTypeFragment = 2
)

var registry = map[string]*ComponentDef{}

// Register adds a component definition to the registry.
func Register(def ComponentDef) {
	registry[def.Tag] = &def
}

// Get returns the component definition for a tag, or nil if not found.
func Get(tag string) *ComponentDef {
	return registry[tag]
}
