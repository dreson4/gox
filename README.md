<p align="center">
  <h1 align="center">GOX</h1>
  <p align="center"><strong>Build native mobile apps with Go.</strong></p>
  <p align="center">Go + JSX-like views. iOS and Android. No VM. No bridge. No JavaScript. Just a compiled binary.</p>
</p>

---

```gox
package main

import "gox"

var styles = gox.Styles{
    "container": gox.Style{Flex: 1, Padding: 24},
    "title":     gox.Style{FontSize: 32, FontWeight: "bold"},
}

view {
    <gox.View style={styles["container"]}>
        <gox.Text style={styles["title"]}>Hello from Go</gox.Text>
    </gox.View>
}
```

```bash
gox run ios
```

That's it. A native mobile app, written in Go, running on your device.

---

## Why Go for mobile?

Go already compiles to iOS and Android. It has goroutines for concurrency, a fast compiler, a great standard library, and produces small static binaries. The only thing missing was a way to build UI.

GOX adds **one thing** to Go: a `view { }` block with JSX-like syntax for declaring native views. Everything else — structs, functions, imports, goroutines, generics, tests — is standard Go.

```
.gox file → GOX compiler → .go file → go build → native binary
```

No interpreter. No virtual DOM. No bridge serialization. Go calls platform APIs directly.

### How GOX compares

| | Language | Runtime overhead | Native views? | Concurrency |
|---|---|---|---|---|
| React Native | JavaScript | Hermes VM + JSON bridge | No — bridge calls | async/await, Promises |
| Flutter | Dart | Skia rendering engine | No — custom renderer | async/await, Futures |
| SwiftUI | Swift | None | Yes | async/await |
| **GOX** | **Go** | **None** | **Yes** | **Goroutines** |

---

## A real example

Here's a screen that fetches data from an API, shows a loading state, and handles errors — the kind of thing every app needs:

```gox
package main

import (
    "gox"
    "net/http"
    "encoding/json"
    "fmt"
)

var posts []Post
var loading bool
var err error

type Post struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
}

// mount runs once when the screen first renders — like Go's init()
func mount() {
    go func() {
        gox.SetState(func() { loading = true })

        resp, e := http.Get("https://jsonplaceholder.typicode.com/posts")
        if e != nil {
            gox.SetState(func() { err = e; loading = false })
            return
        }
        defer resp.Body.Close()

        var data []Post
        json.NewDecoder(resp.Body).Decode(&data)
        gox.SetState(func() { posts = data; loading = false })
    }()
}

view {
    <gox.SafeArea>
        <gox.View style={styles["container"]}>
            <gox.Text style={styles["title"]}>Posts</gox.Text>

            {if loading {
                <gox.Text style={styles["meta"]}>Loading...</gox.Text>
            }}

            {if err != nil {
                <gox.Text style={styles["error"]}>{fmt.Sprintf("Error: %v", err)}</gox.Text>
            }}

            <gox.ScrollView style={styles["scroll"]}>
                {for _, p := range posts {
                    <gox.View style={styles["card"]}>
                        <gox.Text style={styles["postTitle"]}>{p.Title}</gox.Text>
                    </gox.View>
                }}
            </gox.ScrollView>
        </gox.View>
    </gox.SafeArea>
}

var styles = gox.Styles{
    "container":  gox.Style{Flex: 1, Padding: 24, Gap: 16, BackgroundColor: "#F5F5F5"},
    "title":      gox.Style{FontSize: 32, FontWeight: "bold", Color: "#111"},
    "meta":       gox.Style{FontSize: 16, Color: "#666"},
    "error":      gox.Style{FontSize: 16, Color: "#D32F2F"},
    "scroll":     gox.Style{Flex: 1},
    "card":       gox.Style{Padding: 16, BackgroundColor: "#FFF", BorderRadius: 12},
    "postTitle":  gox.Style{FontSize: 18, FontWeight: "600", Color: "#222"},
}
```

Notice what's happening:

- **`func mount()`** — lifecycle hook, like Go's `init()`. Runs once when the screen renders
- **`go func() { ... }()`** — a goroutine, not a Promise or Future
- **`net/http`** — the Go standard library, not a framework HTTP client
- **`encoding/json`** — standard Go JSON decoding
- **`gox.SetState`** — safe from any goroutine, triggers UI update on main thread

You're not learning a new language. You're writing Go.

---

## Goroutines are the killer feature

Every mobile framework invents its own async model. GOX doesn't — it uses Go's.

```gox
// Fetch three APIs concurrently — takes as long as the slowest one
func loadDashboard(ctx context.Context) {
    gox.SetState(func() { loading = true })

    var wg sync.WaitGroup
    wg.Add(3)

    go func() {
        defer wg.Done()
        user, _ = fetchUser(ctx)
    }()

    go func() {
        defer wg.Done()
        posts, _ = fetchPosts(ctx)
    }()

    go func() {
        defer wg.Done()
        notifications, _ = fetchNotifications(ctx)
    }()

    wg.Wait()
    gox.SetState(func() { loading = false })
}
```

No `Promise.all()`. No `async let`. No `Dispatchers.IO`. Just goroutines and a WaitGroup — things every Go developer already knows.

---

## Navigation

```go
// Push a screen
gox.Navigate(profileScreen, gox.Nav("Profile"))

// Go back
gox.GoBack()

// Full control over the nav bar
gox.Navigate(settingsScreen, gox.NavigateOptions{
    Title:       "Settings",
    LargeTitle:  gox.Bool(true),
    HeaderStyle: &gox.HeaderStyle{BackgroundColor: "#1a1a2e"},
    RightButtons: []gox.BarButton{
        {SystemItem: "done", OnPress: func() { save() }},
    },
})
```

---

## Lifecycle

GOX recognizes special function names as lifecycle hooks — just like Go's `init()`:

```gox
// Runs once when the component first renders
func mount() {
    go func() {
        loading = true
        posts, err = FetchPosts(props.UserID)
        loading = false
    }()
}

// Runs when the component is removed from the tree
func unmount() {
    CancelPendingRequests()
}
```

| Function | When it runs |
|---|---|
| `func mount()` | Once, when the component first renders |
| `func unmount()` | Once, when the component is removed |
| `func update()` | After every re-render *(planned)* |
| `func focus()` | When the screen gains navigation focus *(planned)* |
| `func blur()` | When the screen loses navigation focus *(planned)* |

App-level lifecycle is also available:

```go
gox.OnAppForeground(func() { go refreshData() })
gox.OnAppBackground(func() { saveState() })
```

---

## CLI

```bash
# Create a new project
gox init myapp

# Run on iOS simulator
gox run ios

# Run on Android emulator
gox run android

# Run on a specific device
gox run ios --device "iPhone 16 Pro"

# Stream Go logs to terminal
gox run ios --logs

# Production build
gox build ios
gox build android

# Deploy to TestFlight / App Store
gox deploy testflight
gox deploy appstore

# Deploy to Google Play
gox deploy playstore
```

Native projects (Xcode, Gradle) are auto-generated. You never touch them.

---

## Components

GOX maps to native platform views:

| GOX | iOS (UIKit) | Android (planned) | Notes |
|---|---|---|---|
| `View` | `UIView` | `ViewGroup` | Flexbox container (powered by Yoga) |
| `Text` | `UILabel` | `TextView` | Styled text |
| `Button` | `UIButton` | `Button` | Tap handler via `onPress` |
| `TextInput` | `UITextField` | `EditText` | `onChange`, `onSubmit`, `onFocus`, `onBlur` |
| `Image` | `UIImageView` | `ImageView` | URL loading + caching (SDWebImage) |
| `ScrollView` | `UIScrollView` | `ScrollView` | Yoga-computed content size |
| `Switch` | `UISwitch` | `SwitchCompat` | Toggle with `onValueChange` |
| `SafeArea` | Safe area insets | Window insets | Automatic padding for notch/home indicator |

Layout is powered by [Yoga](https://yogalayout.dev/) (the same flexbox engine used by React Native).

---

## How it works under the hood

```
┌─────────────────────────────────────────────┐
│              Your .gox code                  │
├─────────────────────────────────────────────┤
│           GOX Compiler (.gox → .go)          │
├─────────────────────────────────────────────┤
│         Go runtime + GOX framework           │
│   render tree → Yoga layout → frame list     │
├─────────────────────────────────────────────┤
│         cgo bridge (Go ↔ Objective-C)        │
├─────────────────────────────────────────────┤
│    UIKit (iOS) / Android Views (Android)     │
└─────────────────────────────────────────────┘
```

1. The GOX compiler turns `.gox` files into pure `.go` files
2. Your Go code builds a lightweight render tree (`gox.E`, `gox.T`)
3. Yoga computes flexbox layout into flat frames
4. The cgo bridge applies frames to native platform views
5. A hash-based diff engine skips unchanged views on re-render

---

## Status

**GOX is experimental.** It works — we're building apps with it — but APIs will change.

What's working today:
- GOX compiler (lexer, parser, codegen) — 18 tests
- All core components (View, Text, Button, TextInput, Image, ScrollView, Switch)
- Yoga flexbox layout
- Navigation with full nav bar control
- Screen lifecycle (mount, unmount, appear, disappear)
- State management with goroutine-safe `SetState`
- Diff-based re-rendering with frame hashing
- Image caching (SDWebImage)
- `gox run ios` with device picker
- slog logging to terminal via `--logs`

What's coming:
- Android bridge (same Go code, native Android views)
- FlatList (virtualized lists)
- Animations
- `gox init` scaffolding
- `gox deploy` for TestFlight/App Store
- Hot reload
- More components (Modal, Slider, TabBar)

---

## Get involved

GOX is open source. If you've ever wished you could build mobile apps in Go, we'd love your help.

```bash
git clone https://github.com/nicbarker/gox
cd gox
go test ./...
```

See [docs/PLAN.md](docs/PLAN.md) for the full roadmap and [docs/gox-complete-spec.md](docs/gox-complete-spec.md) for the language spec.

---

<p align="center"><em>Write Go. Ship native.</em></p>
