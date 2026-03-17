package codegen

import (
	"gox/internal/compiler/lexer"
	"gox/internal/compiler/parser"
	"strings"
	"testing"
)

func generate(t *testing.T, src string) string {
	t.Helper()
	return generateWithComponent(t, src, false)
}

func generateComponent(t *testing.T, src string, componentName string) string {
	t.Helper()
	l := lexer.New([]byte(src), "test.gox")
	tokens := l.Tokenize()
	p := parser.New(tokens)
	file, errs := p.Parse()
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}
	file.IsComponent = true
	file.ComponentName = componentName
	g := New()
	return g.Generate(file)
}

func generateWithComponent(t *testing.T, src string, isComponent bool) string {
	t.Helper()
	l := lexer.New([]byte(src), "test.gox")
	tokens := l.Tokenize()
	p := parser.New(tokens)
	file, errs := p.Parse()
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}
	if isComponent {
		file.IsComponent = true
		file.ComponentName = "Render" // fallback for old tests
	}
	g := New()
	return g.Generate(file)
}

func TestGenerateHelloWorld(t *testing.T) {
	out := generate(t, `package app

import "gox"

view {
    <gox.View>
        <gox.Text>Hello World</gox.Text>
    </gox.View>
}
`)
	t.Log("Generated:\n" + out)

	// Should contain package declaration
	if !strings.Contains(out, "package app") {
		t.Error("missing package declaration")
	}

	// Should contain render function
	if !strings.Contains(out, "func render()") {
		t.Error("missing render function")
	}

	// Should use gox.E for elements
	if !strings.Contains(out, `gox.E("View"`) {
		t.Error("missing gox.E(\"View\")")
	}
	if !strings.Contains(out, `gox.E("Text"`) {
		t.Error("missing gox.E(\"Text\")")
	}

	// Should contain text
	if !strings.Contains(out, `gox.T("Hello World")`) {
		t.Error("missing text node")
	}
}

func TestGenerateSelfClosing(t *testing.T) {
	out := generate(t, `package app

import "gox"

view {
    <gox.Image src="photo.jpg" />
}
`)
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, `gox.E("Image"`) {
		t.Error("missing Image element")
	}
	if !strings.Contains(out, `"src"`) {
		t.Error("missing src prop")
	}
}

func TestGenerateExprProp(t *testing.T) {
	out := generate(t, `package app

import "gox"

var styles = gox.Styles{
    "title": gox.Style{FontSize: 28},
}

view {
    <gox.Text style={styles["title"]}>Hello</gox.Text>
}
`)
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, `styles["title"]`) {
		t.Error("missing expression prop")
	}
}

func TestGeneratePreservesGoCode(t *testing.T) {
	out := generate(t, `package posts

import (
    "gox"
    "fmt"
)

type Props struct {
    UserID string
}

type State struct {
    count int
}

view {
    <gox.Text>Hello</gox.Text>
}
`)
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, "type Props struct") {
		t.Error("Props struct should be preserved")
	}
	if !strings.Contains(out, "type State struct") {
		t.Error("State struct should be preserved")
	}
}

func TestGenerateComponentRender(t *testing.T) {
	out := generateComponent(t, `package components

import "gox"

type Props struct {
    Author string
    Body   string
}

view {
    <gox.View>
        <gox.Text>{props.Author}</gox.Text>
        <gox.Text>{props.Body}</gox.Text>
        <gox.Children />
    </gox.View>
}
`, "Comment")
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, "func Comment(props CommentProps, children ...*gox.Node) *gox.Node") {
		t.Error("missing Comment function signature with CommentProps")
	}
	if !strings.Contains(out, "gox.Fragment(children...)") {
		t.Error("missing gox.Children expansion")
	}
	if !strings.Contains(out, "props.Author") {
		t.Error("missing props.Author reference")
	}
	// Should NOT contain render()
	if strings.Contains(out, "func render()") {
		t.Error("component should not generate render(), should generate Comment()")
	}
}

func TestGenerateCustomComponentUsage(t *testing.T) {
	out := generate(t, `package main

import (
    "gox"
    "myapp/components"
)

view {
    <components.Comment author="Alice" body="Great post!">
        <gox.Text>Reply</gox.Text>
    </components.Comment>
}
`)
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, "components.Comment(") {
		t.Error("missing components.Comment call")
	}
	if !strings.Contains(out, `components.CommentProps{Author: "Alice", Body: "Great post!"}`) {
		t.Error("missing props struct literal")
	}
	if !strings.Contains(out, `gox.T("Reply")`) {
		t.Error("missing child text node")
	}
	// Should NOT emit gox.E for the component
	if strings.Contains(out, `gox.E("components.Comment"`) {
		t.Error("should not emit gox.E for custom component")
	}
}

func TestGenerateCustomComponentSelfClosing(t *testing.T) {
	out := generate(t, `package main

import "myapp/widgets"

view {
    <widgets.Avatar src="photo.jpg" size={44} />
}
`)
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, "widgets.Avatar(") {
		t.Error("missing widgets.Avatar call")
	}
	if !strings.Contains(out, "widgets.AvatarProps{") {
		t.Error("missing widgets.AvatarProps type")
	}
	if !strings.Contains(out, `Src: "photo.jpg"`) {
		t.Error("missing Src prop")
	}
	if !strings.Contains(out, "Size: 44") {
		t.Error("missing Size prop")
	}
}

func TestGenerateSamePackageComponent(t *testing.T) {
	out := generate(t, `package post

import "gox"

view {
    <gox.View>
        <Avatar src="photo.jpg" size={32} />
    </gox.View>
}
`)
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, "Avatar(AvatarProps{") {
		t.Error("missing same-package component call: Avatar(AvatarProps{...})")
	}
	if !strings.Contains(out, `Src: "photo.jpg"`) {
		t.Error("missing Src prop")
	}
	// Should NOT have a package prefix
	if strings.Contains(out, "post.Avatar") {
		t.Error("should not have package prefix for same-package component")
	}
}

func TestGenerateSpreadProps(t *testing.T) {
	out := generate(t, `package main

import "myapp/comment"

view {
    <comment.Comment {...c} />
}
`)
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, "comment.Comment(comment.CommentProps(c))") {
		t.Error("missing spread props: expected comment.Comment(comment.CommentProps(c))")
	}
}

func TestGenerateChildrenElement(t *testing.T) {
	out := generateComponent(t, `package card

import "gox"

type Props struct {
    Title string
}

view {
    <gox.View>
        <gox.Text>{props.Title}</gox.Text>
        <gox.Children />
    </gox.View>
}
`, "Card")
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, "gox.Fragment(children...)") {
		t.Error("gox.Children should expand to gox.Fragment(children...)")
	}
}

func TestGenerateLifecycleMount(t *testing.T) {
	out := generate(t, `package app

import "gox"
import "context"

func onMount(ctx context.Context) {
    go fetchData(ctx)
}

view {
    <gox.View>
        <gox.Text>Hello</gox.Text>
    </gox.View>
}
`)
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, "gox.UseLifecycle(gox.ScreenCallbacks{") {
		t.Error("missing UseLifecycle call")
	}
	if !strings.Contains(out, "OnMount: onMount,") {
		t.Error("missing OnMount callback")
	}
	if !strings.Contains(out, "_ = ctx") {
		t.Error("missing ctx suppression")
	}
	// Should still have the original function
	if !strings.Contains(out, "func onMount(ctx context.Context)") {
		t.Error("original onMount function should be preserved")
	}
}

func TestGenerateLifecycleMultiple(t *testing.T) {
	out := generate(t, `package app

import "gox"
import "context"

func onMount(ctx context.Context) {}
func onUnmount() {}
func onAppear(ctx context.Context) {}
func onDisappear() {}

view {
    <gox.Text>Hello</gox.Text>
}
`)
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, "OnMount: onMount,") {
		t.Error("missing OnMount")
	}
	if !strings.Contains(out, "OnUnmount: onUnmount,") {
		t.Error("missing OnUnmount")
	}
	if !strings.Contains(out, "OnAppear: onAppear,") {
		t.Error("missing OnAppear")
	}
	if !strings.Contains(out, "OnDisappear: onDisappear,") {
		t.Error("missing OnDisappear")
	}
}

func TestGenerateLifecycleComponent(t *testing.T) {
	out := generateComponent(t, `package thread

import "gox"
import "context"

type Props struct {
    Author string
}

func onMount(ctx context.Context) {}
func onUnmount() {}

view {
    <gox.Text>{props.Author}</gox.Text>
}
`, "Thread")
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, "func Thread(props ThreadProps") {
		t.Error("missing component function")
	}
	if !strings.Contains(out, "gox.UseLifecycle(gox.ScreenCallbacks{") {
		t.Error("missing UseLifecycle in component")
	}
	if !strings.Contains(out, "OnMount: onMount,") {
		t.Error("missing OnMount in component")
	}
	if !strings.Contains(out, "OnUnmount: onUnmount,") {
		t.Error("missing OnUnmount in component")
	}
}

func TestGenerateNoLifecycle(t *testing.T) {
	out := generate(t, `package app

import "gox"

view {
    <gox.Text>Hello</gox.Text>
}
`)
	t.Log("Generated:\n" + out)

	if strings.Contains(out, "UseLifecycle") {
		t.Error("should not emit UseLifecycle when no lifecycle functions are defined")
	}
}

func TestGenerateIfControlFlow(t *testing.T) {
	out := generate(t, `package app

import "gox"

view {
    {if loading {
        <gox.Text>Loading...</gox.Text>
    }}
}
`)
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, "gox.If(loading") {
		t.Error("missing gox.If call")
	}
	if !strings.Contains(out, `gox.T("Loading...")`) {
		t.Error("missing loading text")
	}
}
