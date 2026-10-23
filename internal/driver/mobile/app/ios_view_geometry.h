// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

#ifndef FYNE_IOS_VIEW_GEOMETRY_H
#define FYNE_IOS_VIEW_GEOMETRY_H

#include <limits.h>
#include <math.h>

// GLKView allocates its drawable from view points, not UIScreen.nativeBounds.
static inline int ios_view_pixels(double extent, double scale) {
	if (!isfinite(extent) || !isfinite(scale) || extent <= 0 || scale <= 0 || extent * scale > INT_MAX) {
		return 0;
	}
	return (int)lround(extent * scale);
}

// The keyboard rectangle is already in view coordinates. Floating/undocked
// keyboards must not shrink the entire canvas by their bounding-box height.
static inline double ios_keyboard_bottom_inset(double width, double height,
	double x, double y, double keyboard_width, double keyboard_height) {
	if (!isfinite(width) || !isfinite(height) || !isfinite(x) || !isfinite(y) ||
		!isfinite(keyboard_width) || !isfinite(keyboard_height) || width <= 0 || height <= 0 ||
		keyboard_width <= 0 || keyboard_height <= 0 || x > 0.5 || x + keyboard_width < width - 0.5 ||
		y >= height || y + keyboard_height < height - 0.5) {
		return 0;
	}
	return fmin(height, fmax(0, height - fmax(0, y)));
}

#endif
