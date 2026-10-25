// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

#include "../ios_view_geometry.h"
#include <stdio.h>
#include <stdlib.h>

static void expect(const char *name, double got, double want) {
	if (!isfinite(got) || fabs(got - want) > 0.000001) {
		fprintf(stderr, "%s: got %g, want %g\n", name, got, want);
		exit(1);
	}
}

int main(void) {
	expect("portrait width", ios_view_pixels(390, 3), 1170);
	expect("portrait height", ios_view_pixels(844, 3), 2532);
	expect("landscape width already rotated", ios_view_pixels(844, 3), 2532);
	expect("landscape height already rotated", ios_view_pixels(390, 3), 1170);
	expect("iPad window narrower than screen", ios_view_pixels(600, 2), 1200);
	expect("fractional rendering scale", ios_view_pixels(400, 2.5), 1000);
	expect("fractional bounds", ios_view_pixels(800.25, 2), 1601);
	expect("zero initial bounds", ios_view_pixels(0, 3), 0);
	expect("zero scale", ios_view_pixels(390, 0), 0);
	expect("invalid scale", ios_view_pixels(390, NAN), 0);
	expect("infinite bounds", ios_view_pixels(INFINITY, 3), 0);
	expect("negative bounds", ios_view_pixels(-1, 3), 0);
	expect("overflow", ios_view_pixels(INT_MAX, 2), 0);

	expect("portrait docked", ios_keyboard_bottom_inset(390, 844, 0, 544, 390, 300), 300);
	expect("landscape docked", ios_keyboard_bottom_inset(844, 390, 0, 190, 844, 200), 200);
	expect("split view intersection", ios_keyboard_bottom_inset(600, 768, -100, 468, 1024, 300), 300);
	expect("partial window overlap", ios_keyboard_bottom_inset(600, 600, 0, 500, 600, 300), 100);
	expect("floating iPad", ios_keyboard_bottom_inset(1024, 768, 400, 300, 320, 240), 0);
	expect("floating at bottom", ios_keyboard_bottom_inset(1024, 768, 400, 528, 320, 240), 0);
	expect("undocked full width", ios_keyboard_bottom_inset(1024, 768, 0, 300, 1024, 240), 0);
	expect("keyboard below view", ios_keyboard_bottom_inset(600, 600, 0, 700, 600, 300), 0);
	expect("hidden keyboard", ios_keyboard_bottom_inset(600, 600, 0, 0, 0, 0), 0);
	expect("invalid keyboard", ios_keyboard_bottom_inset(600, 600, 0, NAN, 600, 300), 0);
	expect("keyboard covers whole view", ios_keyboard_bottom_inset(600, 600, 0, -100, 600, 800), 600);
	puts("iOS view geometry: all cases passed");
	return 0;
}
