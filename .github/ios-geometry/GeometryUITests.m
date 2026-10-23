// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

#import <XCTest/XCTest.h>
#include <math.h>

@interface GeometryUITests : XCTestCase
@end

@implementation GeometryUITests
- (NSDictionary *)report:(XCUIElement *)element {
	NSString *value = [element.value isKindOfClass:NSString.class] ? element.value : @"";
	NSData *data = [value dataUsingEncoding:NSUTF8StringEncoding];
	return data == nil ? @{} : [NSJSONSerialization JSONObjectWithData:data options:0 error:nil] ?: @{};
}
- (NSDictionary *)stable:(XCUIElement *)label mode:(NSInteger)mode {
	NSPredicate *ready = [NSPredicate predicateWithBlock:^BOOL(id object, NSDictionary *bindings) {
		NSDictionary *r = [self report:label];
		double scale = [r[@"goScale"] doubleValue];
		double nativeScale = [r[@"viewScale"] doubleValue];
		double width = [r[@"drawableWidth"] doubleValue];
		double height = [r[@"drawableHeight"] doubleValue];
		return r.count != 0 && [r[@"mode"] integerValue] == mode && width > 0 && height > 0 &&
			fabs([r[@"goWidth"] doubleValue] * scale - width) <= 1 &&
			fabs([r[@"goHeight"] doubleValue] * scale - height) <= 1 &&
			fabs([r[@"viewWidth"] doubleValue] * nativeScale - width) <= 1 &&
			fabs([r[@"viewHeight"] doubleValue] * nativeScale - height) <= 1;
	}];
	XCTAssertEqual([XCTWaiter waitForExpectations:@[[[XCTNSPredicateExpectation alloc] initWithPredicate:ready object:label]] timeout:20], XCTWaiterResultCompleted,
		@"Go, UIKit view and GL drawable disagree: %@", label.value);
	NSDictionary *result = [self report:label];
	NSLog(@"FYNE_UIKIT_GEOMETRY %@", label.value);
	return result;
}
- (void)testRotationAndReducedViewportsKeepRenderingAndTouchesAligned {
	self.continueAfterFailure = NO;
	XCUIApplication *app = [[XCUIApplication alloc] initWithBundleIdentifier:@"io.fyne.geometry"];
	[app launch];
	XCUIElement *label = app.staticTexts[@"geometry-report"];
	XCTAssertTrue([label waitForExistenceWithTimeout:30]);
	NSArray *orientations = @[@(UIDeviceOrientationPortrait), @(UIDeviceOrientationLandscapeLeft), @(UIDeviceOrientationPortrait), @(UIDeviceOrientationLandscapeRight)];
	for (NSNumber *orientation in orientations) {
		XCUIDevice.sharedDevice.orientation = orientation.integerValue;
		BOOL landscape = orientation.integerValue == UIDeviceOrientationLandscapeLeft || orientation.integerValue == UIDeviceOrientationLandscapeRight;
		NSPredicate *rotated = [NSPredicate predicateWithBlock:^BOOL(id object, NSDictionary *bindings) {
			return (app.frame.size.width > app.frame.size.height) == landscape;
		}];
		XCTAssertEqual([XCTWaiter waitForExpectations:@[[[XCTNSPredicateExpectation alloc] initWithPredicate:rotated object:app]] timeout:20], XCTWaiterResultCompleted);
		for (NSNumber *mode in @[@0, @1, @2, @0]) {
			[app.segmentedControls[@"viewport-mode"].buttons[@[@"Full", @"Half", @"Small"][mode.integerValue]] tap];
			NSDictionary *r = [self stable:label mode:mode.integerValue];
			NSInteger previous = [r[@"taps"] integerValue];
			double x = [r[@"viewWidth"] doubleValue] * 0.5;
			double y = [r[@"viewHeight"] doubleValue] * 0.65;
			XCUICoordinate *origin = [app coordinateWithNormalizedOffset:CGVectorMake(0, 0)];
			[[origin coordinateWithOffset:CGVectorMake([r[@"frameX"] doubleValue] + x, [r[@"frameY"] doubleValue] + y)] tap];
			NSPredicate *tapped = [NSPredicate predicateWithBlock:^BOOL(id object, NSDictionary *bindings) {
				return [[self report:label][@"taps"] integerValue] > previous;
			}];
			XCTAssertEqual([XCTWaiter waitForExpectations:@[[[XCTNSPredicateExpectation alloc] initWithPredicate:tapped object:label]] timeout:10], XCTWaiterResultCompleted);
			r = [self report:label];
			double fyneScale = [r[@"goScale"] doubleValue];
			double nativeScale = [r[@"viewScale"] doubleValue];
			XCTAssertEqualWithAccuracy([r[@"tapX"] doubleValue] * fyneScale, x * nativeScale, 2);
			// The mobile driver intentionally compensates iOS finger taps by -8 dp.
			XCTAssertEqualWithAccuracy(([r[@"tapY"] doubleValue] + 8) * fyneScale, y * nativeScale, 2);
		}
	}
	XCUIDevice.sharedDevice.orientation = UIDeviceOrientationPortrait;
	[app terminate];
}
@end
