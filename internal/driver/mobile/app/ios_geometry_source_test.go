// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

package app

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// These contracts complement the native UIKit compilation: Linux mobile tests
// cannot compile the Objective-C bridge and used to miss its global-screen path.
func iosGeometrySource(t *testing.T, name string) []byte {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the native source")
	}
	source, err := os.ReadFile(filepath.Join(filepath.Dir(file), name))
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestIOSGeometryUsesTheAttachedView(t *testing.T) {
	text := string(iosGeometrySource(t, "darwin_ios.m"))
	for _, required := range []string{
		"CGRect bounds = self.view.bounds", "CGFloat scale = self.view.contentScaleFactor",
		"scene.effectiveGeometry.interfaceOrientation", "self.view.safeAreaInsets",
		"viewDidLayoutSubviews", "viewSafeAreaInsetsDidChange", "UIKeyboardWillChangeFrameNotification",
		"fromCoordinateSpace:window.screen.coordinateSpace", "ios_keyboard_bottom_inset",
		"ios_view_pixels(bounds.size.width, scale)", "ios_view_pixels(bounds.size.height, scale)",
		"CGFloat scale = view.contentScaleFactor", "locationInView:view",
		"sendTouches(TOUCH_TYPE_CANCEL, touches, self.view)",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("iOS geometry is missing %q", required)
		}
	}
	for _, legacy := range []string{"statusBarOrientation", "nativeScale", "keyboardHeight", "UIApplication.sharedApplication.keyWindow"} {
		if strings.Contains(text, legacy) {
			t.Errorf("iOS geometry still uses global or unconverted %q", legacy)
		}
	}
	// The one remaining use only initializes the physical DisplayMetrics at launch.
	if strings.Count(text, "nativeBounds") > 1 {
		t.Error("window geometry still reads the physical screen bounds")
	}
}

func TestIOSConfigDoesNotRotateAlreadyOrientedPixels(t *testing.T) {
	source, err := parser.ParseFile(token.NewFileSet(), "darwin_ios.go", iosGeometrySource(t, "darwin_ios.go"), 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(source, func(node ast.Node) bool {
		assignment, ok := node.(*ast.AssignStmt)
		if !ok || len(assignment.Lhs) != 2 || len(assignment.Rhs) != 2 {
			return true
		}
		name := func(expression ast.Expr) string {
			if id, ok := expression.(*ast.Ident); ok {
				return id.Name
			}
			return ""
		}
		if name(assignment.Lhs[0]) == "width" && name(assignment.Lhs[1]) == "height" &&
			name(assignment.Rhs[0]) == "height" && name(assignment.Rhs[1]) == "width" {
			t.Error("Go swaps the view pixels a second time in landscape")
		}
		return true
	})
}
