//go:build ios

// Package ios provides the iOS platform integration.
//
// Architecture: Go builds the view tree and exports it as JSON via a C function.
// The Objective-C side (bridge.m) calls GoxGetTree(), parses the JSON,
// and creates UIKit views. This avoids Go needing UIKit headers.
package ios
