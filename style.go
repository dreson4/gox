package gox

// Style defines visual properties for a view element.
// Modeled after CSS flexbox with native mobile conventions.
type Style struct {
	// Dimensions
	Width, Height                            float64
	MinWidth, MinHeight, MaxWidth, MaxHeight float64

	// Flex layout
	Flex           float64
	FlexDirection  string // "row", "column", "row-reverse", "column-reverse"
	JustifyContent string // "start", "center", "end", "between", "around", "evenly"
	AlignItems     string // "start", "center", "end", "stretch", "baseline"
	AlignSelf      string // "auto", "start", "center", "end", "stretch"
	FlexWrap       string // "nowrap", "wrap", "wrap-reverse"
	Gap            float64
	RowGap         float64
	ColumnGap      float64

	// Spacing
	Padding                                                    float64
	PaddingTop, PaddingRight, PaddingBottom, PaddingLeft       float64
	Margin                                                     float64
	MarginTop, MarginRight, MarginBottom, MarginLeft           float64

	// Position
	Position                 string // "relative", "absolute"
	Top, Right, Bottom, Left float64
	ZIndex                   int

	// Appearance
	BackgroundColor string
	Opacity         float64
	Overflow        string // "visible", "hidden", "scroll"

	// Border
	BorderWidth                                                          float64
	BorderTopWidth, BorderRightWidth, BorderBottomWidth, BorderLeftWidth float64
	BorderColor                                                         string
	BorderTopColor, BorderRightColor, BorderBottomColor, BorderLeftColor string
	BorderRadius                                                        float64
	BorderTopLeftRadius, BorderTopRightRadius                            float64
	BorderBottomLeftRadius, BorderBottomRightRadius                      float64

	// Text
	FontSize       float64
	FontWeight     string // "normal", "bold", "100"-"900"
	FontFamily     string
	Color          string
	TextAlign      string // "left", "center", "right", "justify"
	LineHeight     float64
	LetterSpacing  float64
	TextDecoration string // "none", "underline", "line-through"
	TextTransform  string // "none", "uppercase", "lowercase", "capitalize"

	// Transform
	TranslateX, TranslateY float64
	Scale, ScaleX, ScaleY  float64
	Rotation               float64 // degrees

	// Shadow
	ShadowColor              string
	ShadowOffsetX, ShadowOffsetY float64
	ShadowRadius             float64
	ShadowOpacity            float64
}

// Style accessor methods — used by components package via interface assertions.

func (s Style) GetFontSize() float64  { return s.FontSize }
func (s Style) GetFontWeight() string  { return s.FontWeight }
func (s Style) GetColor() string       { return s.Color }
func (s Style) GetHeight() float64     { return s.Height }

// Styles is a named map of styles for convenient grouping.
//
//	var styles = gox.Styles{
//	    "title": gox.Style{FontSize: 24, FontWeight: "bold"},
//	}
type Styles map[string]Style
