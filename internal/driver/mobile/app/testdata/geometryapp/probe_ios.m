// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

#import <UIKit/UIKit.h>
#import <GLKit/GLKit.h>

// Test-only UIKit containment gives the real Fyne controller a changing viewport,
// including an offset split-view-sized region. No production geometry is mocked.
@interface GeometryProbeContainer : UIViewController
@property (strong, nonatomic) UIViewController *content;
@property (strong, nonatomic) UILabel *report;
@property (strong, nonatomic) UISegmentedControl *viewport;
@end

@implementation GeometryProbeContainer
- (void)viewDidLoad {
	// Keep this metrics/touch probe deterministic. XCTest can otherwise spend a
	// minute waiting for a simulator's missing rotation-animation completion.
	// The real controller still receives every layout and orientation change.
	[UIView setAnimationsEnabled:NO];
	[super viewDidLoad];
	[self addChildViewController:self.content];
	[self.view addSubview:self.content.view];
	[self.content didMoveToParentViewController:self];
	self.report = [[UILabel alloc] init];
	self.report.text = @"Geometry probe";
	self.report.accessibilityIdentifier = @"geometry-report";
	self.report.isAccessibilityElement = YES;
	self.report.userInteractionEnabled = NO;
	self.report.backgroundColor = UIColor.whiteColor;
	[self.view addSubview:self.report];
	self.viewport = [[UISegmentedControl alloc] initWithItems:@[@"Full", @"Half", @"Small"]];
	self.viewport.accessibilityIdentifier = @"viewport-mode";
	self.viewport.selectedSegmentIndex = 0;
	[self.viewport addTarget:self action:@selector(changeViewport:) forControlEvents:UIControlEventValueChanged];
	[self.view addSubview:self.viewport];
}
- (void)changeViewport:(id)sender {
	[self.view setNeedsLayout];
	[self.view layoutIfNeeded];
}
- (void)viewDidLayoutSubviews {
	[super viewDidLayoutSubviews];
	CGRect bounds = self.view.bounds;
	CGFloat width = bounds.size.width;
	CGFloat height = bounds.size.height;
	NSInteger mode = self.viewport.selectedSegmentIndex;
	CGSize size = CGSizeMake(width * (mode == 0 ? 1 : mode == 1 ? 0.5 : 0.6), height * (mode == 2 ? 0.5 : 1));
	self.content.view.frame = CGRectMake((width - size.width) / 2, (height - size.height) / 2, size.width, size.height);
	CGFloat top = self.view.safeAreaInsets.top;
	self.report.frame = CGRectMake(5, top + 2, MAX(width - 10, 0), 24);
	self.viewport.frame = CGRectMake(5, top + 28, MAX(width - 10, 0), 32);
}
@end

void geometryProbeUpdate(double width, double height, double scale, int taps, double tapX, double tapY) {
	dispatch_async(dispatch_get_main_queue(), ^{
		id delegate = UIApplication.sharedApplication.delegate;
		UIWindow *window = [delegate window];
		if (window == nil || window.rootViewController == nil) {
			return;
		}
		GeometryProbeContainer *container;
		if ([window.rootViewController isKindOfClass:GeometryProbeContainer.class]) {
			container = (GeometryProbeContainer *)window.rootViewController;
		} else {
			UIViewController *content = window.rootViewController;
			container = [[GeometryProbeContainer alloc] init];
			container.content = content;
			window.rootViewController = container;
			[window makeKeyAndVisible];
		}
		GLKView *view = (GLKView *)container.content.view;
		CGRect frame = [view convertRect:view.bounds toView:window];
		NSDictionary *report = @{
			@"goWidth": @(width), @"goHeight": @(height), @"goScale": @(scale),
			@"viewWidth": @(view.bounds.size.width), @"viewHeight": @(view.bounds.size.height),
			@"frameX": @(frame.origin.x), @"frameY": @(frame.origin.y),
			@"viewScale": @(view.contentScaleFactor),
			@"drawableWidth": @(view.drawableWidth), @"drawableHeight": @(view.drawableHeight),
			@"taps": @(taps), @"tapX": @(tapX), @"tapY": @(tapY),
			@"mode": @(container.viewport.selectedSegmentIndex),
			@"animationsEnabled": @(UIView.areAnimationsEnabled)
		};
		NSData *data = [NSJSONSerialization dataWithJSONObject:report options:0 error:nil];
		container.report.accessibilityValue = [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
	});
}
