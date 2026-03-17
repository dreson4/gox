# GOX — The Go UI Framework

## What Is GOX?

GOX is a native mobile UI framework for Go. It lets you build iOS and Android apps using Go — a language that already compiles to both platforms — by adding a single new concept: a JSX-like `view` block for declaring UI.

Everything else in a `.gox` file is standard Go. Structs, functions, imports, goroutines, generics — it's all Go. The compiler transpiles `.gox` files into pure `.go` files, which then compile to native binaries via `go build`. No runtime, no VM, no bridge.

---

## Why GOX?

| Framework | Language | Runtime | Bundle Size | Build Time | Native? |
|-----------|----------|---------|-------------|------------|---------|
| React Native | JavaScript | Hermes VM + Bridge | 50MB+ | Minutes (Metro) | No — bridge |
| Flutter | Dart | Skia engine | 15MB+ | Seconds-minutes | No — custom renderer |
| Kotlin Multiplatform | Kotlin | JVM (Android) / Native (iOS) | Varies | Minutes (Gradle) | Partial |
| SwiftUI | Swift | Native | Small | Seconds | Yes — Apple only |
| **GOX** | **Go** | **None** | **Small** | **Seconds** | **Yes — both platforms** |

GOX produces a single compiled binary per platform. No interpreter, no virtual DOM, no bridge serialization. Go functions call platform APIs directly. Goroutines handle concurrency without async/await chains. Generics provide type-safe data fetching. The Go ecosystem — testing, profiling, tooling — works out of the box.

---

## Design Decisions

### 1. One new thing, everything else is Go

The `.gox` file format is standard Go with one addition: the `view` block contains JSX-like syntax for declaring UI. Everything outside the view block — structs, functions, variables, imports — is unchanged Go. This means:
- Go formatters, linters, and LSP work on 95% of the file
- No new type system, no decorators, no macros
- A Go developer reads GOX and understands it immediately

### 2. JSX-like syntax for views, not a Go DSL

We considered Go-native alternatives like `ScrollView(style: x) { Text { "hello" } }`. The problem: control flow (`if`, `for`) becomes ambiguous inside nested blocks, and it looks like no language anyone knows. JSX is familiar to millions of developers. The angle brackets create a clear visual boundary between "structure" and "logic." Go control flow inside JSX uses `{if ... { }}` / `{for ... { }}` — outer braces enter Go mode, inner braces are the Go block.

### 3. Explicit imports, no magic globals

Every view comes from a package. Core views use `gox.View`, `gox.Text`, etc. Your components use `posts.PostCard`, `components.Spinner`, etc. No implicit globals, no auto-imports, no component registry. Go's import system is the component system.

We chose `gox.` prefix over dot imports (`. "gox"`) because dot imports collide when multiple packages export the same name — and in a UI framework, that happens immediately (`Button`, `Text`, etc.).

### 4. One way to do everything

- Styles are always a prop. No shorthand dot syntax, no cascading, no inheritance.
- Events are always `func() { }`. No arrow functions, no method references.
- State is always a struct. No hooks, no signals, no observables.
- Navigation is always an explicit router. No file-based magic.
- Animation is always a component. No special language blocks.

### 5. Clean separation: .gox for views, .go for logic

Every component is a package with two files. The `.gox` file handles presentation. The `.go` file handles logic. The `.go` file is pure Go — testable with `go test`, no framework imports required. This separation is encouraged but not enforced; you can put helper functions in the `.gox` file if they're small.

### 6. Goroutines as the concurrency model

No promises, no async/await, no callbacks, no RxJava. You launch a goroutine, you assign to state when you're done. The framework marshals state mutations to the UI thread automatically. Go developers already know goroutines — GOX doesn't invent a new async model.

---

## File Structure

```
myapp/
├── app/
│   ├── app.gox              ← root component
│   ├── app.go               ← router setup
│   └── models/
│       └── post.go           ← data types (pure Go)
├── screens/
│   ├── home/
│   │   ├── home.gox
│   │   └── home.go
│   ├── posts/
│   │   ├── posts.gox
│   │   └── posts.go
│   └── post_detail/
│       ├── post_detail.gox
│       └── post_detail.go
├── components/
│   ├── spinner/
│   │   └── spinner.gox
│   ├── error_banner/
│   │   ├── error_banner.gox
│   │   └── error_banner.go
│   └── empty/
│       └── empty.gox
├── go.mod
└── go.sum
```

Each component is a Go package. A package with a `.gox` file is a renderable component. A package with only `.go` files is a regular Go library.

---

## Language Specification

### The .gox File

A `.gox` file is standard Go with one extension: the `view` block.

```gox
package posts

import (
    "gox"
    "app/components/spinner"
    "app/components/error_banner"
    "app/models"
    "fmt"
)

// Props — standard Go struct
// Defines input from parent component. Read-only.
type Props struct {
    UserID string
}

// State — standard Go struct
// The compiler makes fields reactive: assigning triggers re-render.
type State struct {
    posts   []models.Post
    loading bool
    err     error
    search  string
}

// onMount — standard Go function, special name (like init())
// Runs once when the component first renders.
func onMount(ctx context.Context) {
    go func() {
        loading = true
        posts, err = FetchPosts(ctx, props.UserID)
        loading = false
    }()
}

// onUnmount — standard Go function, special name
// Runs when the component is removed. Clean up here.
func onUnmount() {
    CancelPendingRequests()
}

// Styles — standard Go variable
var styles = gox.Styles{
    "container": gox.Style{Flex: 1, BackgroundColor: "#F5F5F5"},
    "title":     gox.Style{FontSize: 28, FontWeight: "bold", Color: "#111"},
    "subtitle":  gox.Style{FontSize: 14, Color: "#888", MarginTop: 4},
}

// view — THE ONE NON-STANDARD-GO BLOCK
// Contains JSX-like syntax for declaring UI.
view {
    <gox.Screen style={styles["container"]}>
        <gox.Text style={styles["title"]}>My Posts</gox.Text>
        <gox.Text style={styles["subtitle"]}>
            {fmt.Sprintf("%d posts", len(posts))}
        </gox.Text>
    </gox.Screen>
}
```

### What's Standard Go vs What's New

| Feature | Syntax | Standard Go? |
|---------|--------|-------------|
| Package & imports | `package posts`, `import "gox"` | Yes |
| Props | `type Props struct { }` | Yes |
| State | `type State struct { }` | Yes (compiler adds reactivity) |
| Styles | `var styles = gox.Styles{ }` | Yes |
| Mount lifecycle | `func onMount(ctx context.Context) { }` | Yes (special name, like `init()`) |
| Unmount lifecycle | `func onUnmount() { }` | Yes (special name) |
| Helper functions | `func doSomething() { }` | Yes |
| Goroutines | `go func() { }()` | Yes |
| **View template** | **`view { <JSX /> }`** | **No — the one new thing** |

---

### Props

Standard Go struct named `Props`. The compiler recognizes this type name and makes it available as `props` within the component.

```gox
type Props struct {
    UserID   string
    Title    string
    OnSelect func(id string)
    Style    gox.Style
}
```

Access via `props.FieldName`:

```gox
func onMount(ctx context.Context) {
    go func() {
        data, err = FetchData(ctx, props.UserID)
    }()
}

view {
    <gox.Text>{props.Title}</gox.Text>
}
```

When using a component, props are passed as JSX attributes:

```gox
<posts.Posts userID="user_123" title="Recent Posts" />
```

---

### State

Standard Go struct named `State`. The compiler makes its fields reactive — assigning to a field triggers a re-render. State fields are accessed directly by name within the component (not `state.loading`, just `loading`).

```gox
type State struct {
    count   int
    items   []models.Item
    loading bool
    err     error
}
```

Modify by assignment anywhere — in `onMount()`, in event handlers, in goroutines, in functions called from the `.go` file:

```gox
func onMount(ctx context.Context) {
    go func() {
        loading = true
        items, err = FetchItems(ctx)
        loading = false
    }()
}
```

```gox
// In the view — event handler
<gox.Button onPress={func() { count = count + 1 }}>
    <gox.Text>Count: {count}</gox.Text>
</gox.Button>
```

The framework ensures state mutations from any goroutine are marshalled to the UI thread.

---

### Lifecycle Functions

Like Go's `init()`, GOX recognizes specific function names as lifecycle hooks. The compiler detects them and wires them up automatically.

| Function | When it runs |
|----------|-------------|
| `func onMount(ctx context.Context)` | Once, when the screen first renders |
| `func onUnmount()` | Once, when the screen is removed from the tree |
| `func onAppear(ctx context.Context)` | Each time the screen becomes visible |
| `func onDisappear()` | Each time the screen is hidden |

The `ctx` is a per-screen context that's automatically cancelled when the screen is popped — goroutines clean up without manual bookkeeping.

```gox
var timer *gox.Timer

func onMount(ctx context.Context) {
    // Fetch initial data
    go func() {
        loading = true
        posts, err = FetchPosts(ctx, props.UserID)
        loading = false
    }()

    // Poll for updates every 30 seconds
    timer = gox.Every(30*time.Second, func() {
        posts, _ = FetchPosts(ctx, props.UserID)
    })
}

func onUnmount() {
    timer.Stop()
    CancelPendingRequests()
}

func onAppear(ctx context.Context) {
    go refreshData(ctx)
}

func onDisappear() {
    pausePolling()
}
```

---

### View Block

The `view` block is the only non-standard-Go syntax. It contains JSX-like elements with Go control flow embedded via `{ }`.

#### Elements

```gox
// With props and children
<gox.ScrollView style={styles["container"]}>
    <gox.Text>Hello</gox.Text>
</gox.ScrollView>

// Self-closing (no children)
<gox.Image src={user.Avatar} style={styles["avatar"]} />

// No props
<gox.View>
    <gox.Text>Content</gox.Text>
</gox.View>
```

#### Props on Elements

`name={expression}` for dynamic values. Quoted strings for literals.

```gox
<gox.Text style={styles["title"]}>{user.Name}</gox.Text>
<gox.TextInput placeholder="Search..." />
<gox.Button onPress={func() { HandleSubmit() }}>
    <gox.Text>Submit</gox.Text>
</gox.Button>

// Multi-line props
<gox.TextInput
    value={search}
    placeholder="Search posts..."
    onChange={func(v string) { search = v }}
    style={styles["input"]}
/>
```

#### Text Content and Expressions

`{expression}` evaluates any Go expression:

```gox
<gox.Text>Hello World</gox.Text>
<gox.Text>Hello {user.Name}</gox.Text>
<gox.Text>{fmt.Sprintf("%d items", len(items))}</gox.Text>
<gox.Text>{err.Error()}</gox.Text>
```

#### Go Control Flow

`if`, `for`, `switch` are wrapped in `{ }` to distinguish from text content:

```gox
view {
    <gox.View>

        {if loading {
            <spinner.Spinner size="large" />
        }}

        {if err != nil {
            <gox.Text style={styles["error"]}>{err.Error()}</gox.Text>
        }}

        {for _, post := range posts {
            <gox.View key={post.ID} style={styles["card"]}>
                <gox.Text style={styles["cardTitle"]}>{post.Title}</gox.Text>
                <gox.Text style={styles["cardBody"]}>{post.Body}</gox.Text>
            </gox.View>
        }}

        {if len(posts) == 0 && !loading {
            <gox.Text>No posts yet</gox.Text>
        }}

        {switch user.Role {
        case "admin":
            <AdminPanel />
        case "moderator":
            <ModTools />
        default:
            <UserView />
        }}

    </gox.View>
}
```

**Rules:**
- `{if condition { ... }}` — outer `{ }` enters Go mode, inner `{ }` is the block
- `{for ... range ... { ... }}` — same pattern
- `{switch expr { case: ... }}` — same pattern
- No `return` keyword — JSX inside blocks renders in place
- Fully nestable — control flow inside elements, elements inside control flow

---

### Styles

Styles are standard Go data. The framework provides the `gox.Style` struct and `gox.Styles` map type. Style is always passed as a prop — one way, no shortcuts.

#### Defining Styles

```gox
// Map style — convenient for many styles
var styles = gox.Styles{
    "container":  gox.Style{Flex: 1, Padding: 16, BackgroundColor: "#F5F5F5"},
    "title":      gox.Style{FontSize: 24, FontWeight: "bold", Color: "#111"},
    "subtitle":   gox.Style{FontSize: 14, Color: "#888"},
}

// Individual variables — also fine
var headerStyle = gox.Style{Padding: 16, BackgroundColor: "#FFF"}
```

#### Using Styles

```gox
<gox.View style={styles["container"]}>
    <gox.Text style={styles["title"]}>Hello</gox.Text>
</gox.View>
```

#### Combining Styles

```gox
<gox.Text style={gox.Merge(styles["title"], styles["highlighted"])}>
    Important
</gox.Text>
```

#### Dynamic Styles

Compute in Go — inline or in your `.go` file:

```gox
// Inline with conditional helper
<gox.View style={gox.Merge(styles["card"], gox.Style{
    BackgroundColor: StatusColor(post.Status),
    Opacity:         gox.When(isActive, 1.0, 0.5),
})}>
    <gox.Text>{post.Title}</gox.Text>
</gox.View>

// Function from .go file
<gox.View style={CardStyle(isActive, itemCount)}>
    <gox.Text>Content</gox.Text>
</gox.View>
```

```go
// posts.go
func CardStyle(active bool, count int) gox.Style {
    s := gox.Style{Padding: 16, BorderRadius: 8}
    if active {
        s.BackgroundColor = "#E3F2FD"
    }
    if count > 10 {
        s.BorderColor = "#FF9800"
        s.BorderWidth = 2
    }
    return s
}
```

---

### Components

Components are Go packages. Any package with a `.gox` file containing a `view` block is a renderable component.

#### Using Components

```gox
import (
    "gox"
    "app/screens/posts"
    "app/components/spinner"
)

view {
    <posts.Posts userID="user_123" />
    <spinner.Spinner size="large" />
}
```

#### The `key` Prop

For list identity tracking — helps the framework efficiently update, reorder, and animate list items:

```gox
{for _, item := range items {
    <ListItem key={item.ID} item={item} />
}}
```

#### Children

Components receive children from their parent. Render them with `<gox.Children />`:

```gox
// card/card.gox
package card

import "gox"

type Props struct {
    Title string
}

var styles = gox.Styles{
    "wrapper": gox.Style{Padding: 16, BorderRadius: 12, BackgroundColor: "#FFF"},
    "title":   gox.Style{FontSize: 18, FontWeight: "bold", MarginBottom: 8},
}

view {
    <gox.View style={styles["wrapper"]}>
        <gox.Text style={styles["title"]}>{props.Title}</gox.Text>
        <gox.Children />
    </gox.View>
}
```

Usage:

```gox
<card.Card title="My Card">
    <gox.Text>This renders where Children appears</gox.Text>
    <gox.Image src="photo.jpg" />
</card.Card>
```

#### Fragment

Group elements without adding a wrapper view to the tree:

```gox
<gox.Fragment>
    <gox.Text>One</gox.Text>
    <gox.Text>Two</gox.Text>
</gox.Fragment>
```

---

### Animations

Animation is a component (`gox.Animated`), not a language feature. It wraps a child and interpolates a style property over time when the target value changes.

#### Basic Animation

```gox
<gox.Animated
    prop="opacity"
    to={gox.When(visible, 1.0, 0.0)}
    duration={300}
    easing="easeInOut"
>
    <gox.View style={styles["card"]}>
        <gox.Text>I fade in and out</gox.Text>
    </gox.View>
</gox.Animated>
```

When `visible` changes, `opacity` transitions smoothly over 300ms.

#### Multiple Properties

Nest `Animated` components:

```gox
<gox.Animated prop="opacity" to={gox.When(expanded, 1.0, 0.0)} duration={300}>
    <gox.Animated prop="height" to={gox.When(expanded, 400.0, 0.0)} duration={250} easing="spring">
        <gox.View>
            <gox.Text>Expandable content</gox.Text>
        </gox.View>
    </gox.Animated>
</gox.Animated>
```

#### Animatable Properties

`opacity`, `translateX`, `translateY`, `scale`, `scaleX`, `scaleY`, `rotation`, `width`, `height`, `borderRadius`, `backgroundColor`

#### Easing Options

`"linear"`, `"easeIn"`, `"easeOut"`, `"easeInOut"`, `"spring"`

#### Sequencing

Orchestrate in your `.go` file:

```go
func ShowWithBounce() {
    visible = true
    gox.After(150*time.Millisecond, func() {
        expanded = true
    })
}
```

Or use the `onDone` callback:

```gox
<gox.Animated
    prop="opacity"
    to={1.0}
    duration={200}
    onDone={func() { expanded = true }}
>
    <gox.View>...</gox.View>
</gox.Animated>
```

---

### Navigation

Explicit router configuration — no file-based magic.

#### Router Setup

```go
// app/app.go
package app

import (
    "gox"
    "app/screens/home"
    "app/screens/posts"
    "app/screens/post_detail"
    "app/screens/settings"
)

var Router = gox.NewRouter(
    gox.Route("/",             home.Screen),
    gox.Route("/posts",        posts.Screen),
    gox.Route("/posts/:id",    post_detail.Screen),
    gox.Route("/settings",     settings.Screen),
)
```

#### Root Component

```gox
// app/app.gox
package app

import "gox"

view {
    <gox.Navigator router={Router} />
}
```

#### Navigation Functions

```go
gox.Navigate("/posts")
gox.Navigate("/posts/" + postID)
gox.GoBack()
gox.GoToRoot()
gox.Replace("/login")     // replace current screen, no back
```

In a view:

```gox
<gox.Button onPress={func() { gox.Navigate("/posts/" + post.ID) }}>
    <gox.Text>View Post</gox.Text>
</gox.Button>
```

#### Route Params

URL params are populated into `Props`. The field name matches the param name:

```go
// Route: "/posts/:id"
// post_detail/post_detail.gox

type Props struct {
    ID string    // populated from :id
}

func onMount(ctx context.Context) {
    go func() {
        post, err = FetchPost(ctx, props.ID)
    }()
}
```

#### Tab Navigation

```go
var Router = gox.NewTabRouter(
    gox.Tab("/home",    home.Screen,    gox.TabOptions{Icon: "home",   Title: "Home"}),
    gox.Tab("/search",  search.Screen,  gox.TabOptions{Icon: "search", Title: "Search"}),
    gox.Tab("/profile", profile.Screen, gox.TabOptions{Icon: "person", Title: "Profile"}),
)
```

Each tab can have its own stack of screens.

#### Route Options

```go
gox.RouteWithOptions("/settings", settings.Screen, gox.RouteOptions{
    Title:      "Settings",
    Transition: "modal",       // "push", "modal", "fade", "none"
    Protected:  true,          // requires auth
})

// Set auth guard globally
gox.SetAuthGuard(
    func() bool { return currentUser != nil },
    "/login",    // redirect here if guard fails
)
```

---

## Core Package API

### `package gox`

One import gives you everything for a typical screen.

```gox
import "gox"
```

#### Primitive Views

**Layout:**

| View | Description | iOS | Android |
|------|-------------|-----|---------|
| `gox.View` | Generic container | UIView | android.view.View |
| `gox.ScrollView` | Scrollable container | UIScrollView | ScrollView |
| `gox.Row` | Horizontal flex container | UIStackView (h) | LinearLayout (h) |
| `gox.Column` | Vertical flex container | UIStackView (v) | LinearLayout (v) |
| `gox.Screen` | Full-screen root, handles safe areas | UIViewController root | Activity root |
| `gox.SafeArea` | Insets for notches / home indicator | SafeAreaView | WindowInsetsCompat |

**Text & Input:**

| View | Description | iOS | Android |
|------|-------------|-----|---------|
| `gox.Text` | Renders text | UILabel | TextView |
| `gox.TextInput` | Editable text field | UITextField | EditText |

**Interactive:**

| View | Description | iOS | Android |
|------|-------------|-----|---------|
| `gox.Button` | Tappable with press feedback | UIButton | MaterialButton |
| `gox.Pressable` | Generic pressable wrapper | UIControl | View + click listener |
| `gox.Switch` | Toggle on/off | UISwitch | SwitchCompat |
| `gox.Slider` | Draggable value selector | UISlider | SeekBar |

**Media:**

| View | Description | iOS | Android |
|------|-------------|-----|---------|
| `gox.Image` | Displays images (URL, asset, bytes) | UIImageView | ImageView |

**Lists:**

| View | Description | iOS | Android |
|------|-------------|-----|---------|
| `gox.FlatList` | Virtualized scrolling list | UICollectionView | RecyclerView |

**Overlay:**

| View | Description | iOS | Android |
|------|-------------|-----|---------|
| `gox.Modal` | Content above current screen | present VC | Dialog |

**Composition:**

| View | Description |
|------|-------------|
| `gox.Children` | Renders children passed to a component |
| `gox.Fragment` | Groups elements without a wrapper view |
| `gox.Animated` | Animates a style property on its child |
| `gox.Navigator` | Renders the current route |

---

#### View Props Reference

```go
// gox.View
type ViewProps struct {
    Style   Style
    ID      string
    OnPress func()
}

// gox.ScrollView
type ScrollViewProps struct {
    Style                Style
    Horizontal           bool
    ShowsScrollIndicator bool
    OnScroll             func(offset float64)
    OnScrollEnd          func(offset float64)
    RefreshControl       *RefreshControl
}

type RefreshControl struct {
    Refreshing bool
    OnRefresh  func()
}

// gox.Text
type TextProps struct {
    Style      Style
    Selectable bool
    MaxLines   int
    OnPress    func()
}

// gox.TextInput
type TextInputProps struct {
    Style        Style
    Value        string
    Placeholder  string
    OnChange     func(value string)
    OnSubmit     func(value string)
    OnFocus      func()
    OnBlur       func()
    Secure       bool
    KeyboardType string    // "default", "email", "numeric", "phone", "url"
    AutoFocus    bool
    MaxLength    int
    MultiLine    bool
}

// gox.Button
type ButtonProps struct {
    Style    Style
    OnPress  func()
    Disabled bool
}

// gox.Switch
type SwitchProps struct {
    Style    Style
    Value    bool
    OnChange func(value bool)
    Disabled bool
}

// gox.Slider
type SliderProps struct {
    Style    Style
    Value    float64
    Min      float64
    Max      float64
    Step     float64
    OnChange func(value float64)
    Disabled bool
}

// gox.Image
type ImageProps struct {
    Style       Style
    Src         string
    Data        []byte
    ContentMode string    // "cover", "contain", "stretch", "center"
    OnLoad      func()
    OnError     func(err error)
    Placeholder string
}

// gox.FlatList — generic
type FlatListProps[T any] struct {
    Style          Style
    Data           []T
    KeyExtractor   func(item T) string
    RenderItem     func(item T, index int)
    OnEndReached   func()
    OnEndThreshold float64
    ItemSeparator  func()
    ListHeader     func()
    ListFooter     func()
    ListEmpty      func()
    Horizontal     bool
    RefreshControl *RefreshControl
}

// gox.Modal
type ModalProps struct {
    Visible   bool
    OnClose   func()
    Animation string    // "slide", "fade", "none"
}

// gox.Animated
type AnimatedProps struct {
    Prop     string
    To       float64
    Duration int
    Delay    int
    Easing   string    // "linear", "easeIn", "easeOut", "easeInOut", "spring"
    OnDone   func()
}
```

---

#### Style Type

```go
type Style struct {
    // Dimensions
    Width, Height                              float64
    MinWidth, MinHeight, MaxWidth, MaxHeight   float64

    // Flex
    Flex           float64
    FlexDirection  string    // "row", "column", "row-reverse", "column-reverse"
    JustifyContent string    // "start", "center", "end", "between", "around", "evenly"
    AlignItems     string    // "start", "center", "end", "stretch", "baseline"
    AlignSelf      string    // "auto", "start", "center", "end", "stretch"
    FlexWrap       string    // "nowrap", "wrap", "wrap-reverse"
    Gap, RowGap, ColumnGap   float64

    // Spacing
    Padding, PaddingTop, PaddingRight, PaddingBottom, PaddingLeft       float64
    Margin, MarginTop, MarginRight, MarginBottom, MarginLeft            float64

    // Position
    Position                    string    // "relative", "absolute"
    Top, Right, Bottom, Left    float64
    ZIndex                      int

    // Appearance
    BackgroundColor string
    Opacity         float64
    Overflow        string    // "visible", "hidden", "scroll"

    // Border
    BorderWidth, BorderTopWidth, BorderRightWidth, BorderBottomWidth, BorderLeftWidth   float64
    BorderColor, BorderTopColor, BorderRightColor, BorderBottomColor, BorderLeftColor   string
    BorderRadius                                                                        float64
    BorderTopLeftRadius, BorderTopRightRadius, BorderBottomLeftRadius, BorderBottomRightRadius float64

    // Text
    FontSize       float64
    FontWeight     string    // "normal", "bold", "100"-"900"
    FontFamily     string
    Color          string
    TextAlign      string    // "left", "center", "right", "justify"
    LineHeight     float64
    LetterSpacing  float64
    TextDecoration string    // "none", "underline", "line-through"
    TextTransform  string    // "none", "uppercase", "lowercase", "capitalize"

    // Transform
    TranslateX, TranslateY float64
    Scale, ScaleX, ScaleY  float64
    Rotation               float64    // degrees

    // Shadow
    ShadowColor              string
    ShadowOffsetX, ShadowOffsetY float64
    ShadowRadius             float64
    ShadowOpacity            float64
}

type Styles map[string]Style
```

---

#### Utility Functions

```go
// Merge multiple styles — later values override earlier
func Merge(styles ...Style) Style

// Conditional value — generic
func When[T any](condition bool, ifTrue T, ifFalse T) T

// Run function after delay
func After(duration time.Duration, fn func())

// Run function on interval — returns stoppable timer
func Every(interval time.Duration, fn func()) *Timer
type Timer struct{}
func (t *Timer) Stop()

// Toast notification
func ShowToast(message string, opts ...ToastOption)
type ToastOption struct {
    Duration int       // milliseconds, default 2000
    Position string    // "top", "bottom", "center"
}

// Logging (stripped from release builds)
func Log(args ...any)
func Logf(format string, args ...any)
```

---

#### Navigation Functions

```go
func NewRouter(routes ...RouteConfig) *Router
func Route(path string, component Component) RouteConfig
func RouteWithOptions(path string, component Component, opts RouteOptions) RouteConfig
func NewTabRouter(tabs ...TabConfig) *Router
func Tab(path string, component Component, opts TabOptions) TabConfig

type RouteOptions struct {
    Title      string
    Transition string    // "push", "modal", "fade", "none"
    Protected  bool
}

type TabOptions struct {
    Icon  string
    Title string
}

func Navigate(path string)
func NavigateWithParams(path string, params map[string]any)
func GoBack()
func GoToRoot()
func Replace(path string)
func SetAuthGuard(check func() bool, redirect string)
```

---

#### Platform Functions

```go
func Platform() string    // "ios", "android"
func OpenURL(url string) error
func CopyToClipboard(text string)
func ReadClipboard() string
func Haptic(style string)         // "light", "medium", "heavy", "success", "error", "warning"
func SetStatusBar(style string)   // "light", "dark", "auto"
func DismissKeyboard()
```

---

#### Storage

```go
func Store(key string, value any) error
func Load(key string, dest any) error
func Delete(key string) error
```

Maps to UserDefaults (iOS) and SharedPreferences (Android).

---

### `package gox/http`

Typed HTTP client with generics. JSON serialization built in.

```go
import "gox/http"

// Typed requests — response is deserialized into T
func Get[T any](url string) (T, error)
func Post[T any](url string, body any) (T, error)
func Put[T any](url string, body any) (T, error)
func Patch[T any](url string, body any) (T, error)
func Del[T any](url string) (T, error)

// Full request control
func Do[T any](req Request) (T, error)

type Request struct {
    Method  string
    URL     string
    Headers map[string]string
    Body    any
    Timeout time.Duration
}

// Configuration
func SetBaseURL(url string)
func SetHeader(key, value string)        // global default header
func SetTimeout(duration time.Duration)  // global default timeout
```

Usage:

```go
// posts.go
import "gox/http"

func FetchPosts(userID string) ([]models.Post, error) {
    return http.Get[[]models.Post]("/users/" + userID + "/posts")
}

func CreatePost(p models.NewPost) (models.Post, error) {
    return http.Post[models.Post]("/posts", p)
}

func DeletePost(id string) error {
    _, err := http.Del[struct{}]("/posts/" + id)
    return err
}
```

---

## Compiler Pipeline

```
.gox source file
  │
  ▼
GOX Parser
  │  Finds the view { } block
  │  Everything else passes through as-is (it's valid Go)
  │
  ▼
View Transformation
  │  JSX elements → Go function calls: gox.Render("ScrollView", props, children...)
  │  {if/for/switch} → standard Go control flow wrapping render calls
  │  {expressions} → Go expressions
  │
  ▼
State Detection
  │  type State struct → reactive wrapper with setter methods
  │  Assignments to state fields → setter calls that trigger re-render
  │
  ▼
Lifecycle Detection
  │  onMount(ctx) → register with component mount lifecycle
  │  onUnmount() → register with component cleanup lifecycle
  │  onAppear(ctx) → register with screen appear lifecycle
  │  onDisappear() → register with screen disappear lifecycle
  │
  ▼
Emits pure .go file
  │  The output is standard Go that imports the gox runtime
  │
  ▼
go build (GOOS=android / GOOS=darwin)
  │
  ▼
Native binary
  │  Go code calling platform APIs directly
  │  No VM, no bridge, no interpreter
```

---

## Full Example: Social Feed App

### `app/models/post.go`

```go
package models

import "time"

type User struct {
    ID     string
    Name   string
    Avatar string
}

type Post struct {
    ID        string
    Author    User
    Title     string
    Body      string
    Image     string
    Likes     int
    Liked     bool
    CreatedAt time.Time
}
```

### `app/app.go`

```go
package app

import (
    "gox"
    "app/screens/home"
    "app/screens/posts"
    "app/screens/post_detail"
    "app/screens/profile"
)

var Router = gox.NewTabRouter(
    gox.Tab("/home", home.Screen, gox.TabOptions{
        Icon: "home", Title: "Home",
    }),
    gox.Tab("/posts", posts.Screen, gox.TabOptions{
        Icon: "list", Title: "Posts",
    }),
    gox.Tab("/profile", profile.Screen, gox.TabOptions{
        Icon: "person", Title: "Profile",
    }),
)

func init() {
    // Add stack routes within tabs
    Router.AddRoute(gox.Route("/posts/:id", post_detail.Screen))
}
```

### `app/app.gox`

```gox
package app

import "gox"

view {
    <gox.Navigator router={Router} />
}
```

### `screens/posts/posts.go`

```go
package posts

import (
    "gox/http"
    "app/models"
    "strings"
)

func FetchPosts(userID string) ([]models.Post, error) {
    return http.Get[[]models.Post]("/users/" + userID + "/posts")
}

func HandleLike(postID string) error {
    return http.Post[struct{}]("/posts/"+postID+"/like", nil)
}

func HandleShare(postID string) error {
    return http.Post[struct{}]("/posts/"+postID+"/share", nil)
}

func FilterPosts(posts []models.Post, query string) []models.Post {
    if query == "" {
        return posts
    }
    var filtered []models.Post
    for _, p := range posts {
        if strings.Contains(
            strings.ToLower(p.Title),
            strings.ToLower(query),
        ) {
            filtered = append(filtered, p)
        }
    }
    return filtered
}

func CancelPendingRequests() {
    // cancel in-flight HTTP requests
}
```

### `screens/posts/posts.gox`

```gox
package posts

import (
    "gox"
    "app/components/spinner"
    "app/components/error_banner"
    "app/components/empty"
    "app/models"
    "fmt"
)

type Props struct {
    UserID string
}

type State struct {
    posts   []models.Post
    loading bool
    err     error
    search  string
}

func onMount(ctx context.Context) {
    go func() {
        loading = true
        posts, err = FetchPosts(ctx, props.UserID)
        loading = false
    }()
}

func onUnmount() {
    CancelPendingRequests()
}

var styles = gox.Styles{
    "container":  gox.Style{Flex: 1, BackgroundColor: "#F5F5F5"},
    "header":     gox.Style{Padding: 16, PaddingBottom: 8},
    "title":      gox.Style{FontSize: 28, FontWeight: "bold", Color: "#111"},
    "subtitle":   gox.Style{FontSize: 14, Color: "#888", MarginTop: 4},
    "input":      gox.Style{Margin: 16, MarginTop: 8, Padding: 12, BorderRadius: 8, BackgroundColor: "#FFF"},
    "card":       gox.Style{Margin: 16, MarginTop: 8, Padding: 16, BorderRadius: 12, BackgroundColor: "#FFF"},
    "cardHeader": gox.Style{FlexDirection: "row", AlignItems: "center"},
    "avatar":     gox.Style{Width: 40, Height: 40, BorderRadius: 20},
    "meta":       gox.Style{MarginLeft: 12},
    "author":     gox.Style{FontSize: 16, FontWeight: "600"},
    "date":       gox.Style{FontSize: 12, Color: "#888"},
    "body":       gox.Style{FontSize: 14, MarginTop: 12, LineHeight: 20, Color: "#333"},
    "postImage":  gox.Style{Width: -1, Height: 200, BorderRadius: 8, MarginTop: 12},
    "actions":    gox.Style{FlexDirection: "row", MarginTop: 12, Gap: 8},
    "likeActive": gox.Style{Color: "#E91E63"},
    "error":      gox.Style{Color: "#FF3B30", Padding: 16},
}

view {
    <gox.Screen style={styles["container"]}>
        <gox.View style={styles["header"]}>
            <gox.Text style={styles["title"]}>My Posts</gox.Text>
            <gox.Text style={styles["subtitle"]}>
                {fmt.Sprintf("%d posts", len(posts))}
            </gox.Text>
        </gox.View>

        <gox.TextInput
            value={search}
            placeholder="Search posts..."
            onChange={func(v string) { search = v }}
            style={styles["input"]}
        />

        {if loading {
            <spinner.Spinner size="large" />
        }}

        {if err != nil {
            <error_banner.ErrorBanner
                message={err.Error()}
                onRetry={func() {
                    go func() {
                        loading = true
                        posts, err = FetchPosts(props.UserID)
                        loading = false
                    }()
                }}
            />
        }}

        <gox.ScrollView>
            {for _, post := range FilterPosts(posts, search) {
                <gox.View key={post.ID} style={styles["card"]}>
                    <gox.View style={styles["cardHeader"]}>
                        <gox.Image
                            src={post.Author.Avatar}
                            style={styles["avatar"]}
                            contentMode="cover"
                        />
                        <gox.View style={styles["meta"]}>
                            <gox.Text style={styles["author"]}>
                                {post.Author.Name}
                            </gox.Text>
                            <gox.Text style={styles["date"]}>
                                {post.CreatedAt.Format("Jan 2")}
                            </gox.Text>
                        </gox.View>
                    </gox.View>

                    <gox.Text style={styles["body"]}>{post.Body}</gox.Text>

                    {if post.Image != "" {
                        <gox.Image
                            src={post.Image}
                            style={styles["postImage"]}
                            contentMode="cover"
                        />
                    }}

                    <gox.View style={styles["actions"]}>
                        <gox.Button onPress={func() {
                            go func() { HandleLike(post.ID) }()
                        }}>
                            <gox.Text style={gox.When(post.Liked, styles["likeActive"], gox.Style{})}>
                                {fmt.Sprintf("Like (%d)", post.Likes)}
                            </gox.Text>
                        </gox.Button>
                        <gox.Button onPress={func() {
                            go func() { HandleShare(post.ID) }()
                        }}>
                            <gox.Text>Share</gox.Text>
                        </gox.Button>
                        <gox.Button onPress={func() {
                            gox.Navigate("/posts/" + post.ID)
                        }}>
                            <gox.Text>Open</gox.Text>
                        </gox.Button>
                    </gox.View>
                </gox.View>
            }}

            {if len(posts) == 0 && !loading {
                <empty.Empty message="No posts yet" />
            }}
        </gox.ScrollView>
    </gox.Screen>
}
```

### `components/spinner/spinner.gox`

```gox
package spinner

import "gox"

type Props struct {
    Size string    // "small", "large"
}

var styles = gox.Styles{
    "container": gox.Style{
        Flex: 1, JustifyContent: "center", AlignItems: "center", Padding: 32,
    },
}

view {
    <gox.View style={styles["container"]}>
        <gox.ActivityIndicator size={props.Size} />
    </gox.View>
}
```

### `components/error_banner/error_banner.gox`

```gox
package error_banner

import "gox"

type Props struct {
    Message string
    OnRetry func()
}

var styles = gox.Styles{
    "container": gox.Style{
        Margin: 16, Padding: 16, BorderRadius: 8,
        BackgroundColor: "#FFF0F0", FlexDirection: "row",
        AlignItems: "center", JustifyContent: "between",
    },
    "message": gox.Style{Color: "#D32F2F", FontSize: 14, Flex: 1},
    "retry":   gox.Style{Color: "#1976D2", FontWeight: "bold", FontSize: 14},
}

view {
    <gox.View style={styles["container"]}>
        <gox.Text style={styles["message"]}>{props.Message}</gox.Text>
        {if props.OnRetry != nil {
            <gox.Button onPress={props.OnRetry}>
                <gox.Text style={styles["retry"]}>Retry</gox.Text>
            </gox.Button>
        }}
    </gox.View>
}
```

### `components/empty/empty.gox`

```gox
package empty

import "gox"

type Props struct {
    Message string
}

var styles = gox.Styles{
    "container": gox.Style{
        Flex: 1, JustifyContent: "center", AlignItems: "center", Padding: 48,
    },
    "text": gox.Style{FontSize: 16, Color: "#999"},
}

view {
    <gox.View style={styles["container"]}>
        <gox.Text style={styles["text"]}>{props.Message}</gox.Text>
    </gox.View>
}
```
