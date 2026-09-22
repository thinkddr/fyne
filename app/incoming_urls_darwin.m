// Copyright (C) 2026 Sytue.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//go:build darwin && !ios && !mobile && !ci && !wasm && !test_web_driver

#include "_cgo_export.h"
#import <AppKit/AppKit.h>
#import <Carbon/Carbon.h>

@interface FyneIncomingURLHandler : NSObject
@end

@implementation FyneIncomingURLHandler
- (void)handleURL:(NSAppleEventDescriptor *)event withReplyEvent:(NSAppleEventDescriptor *)reply {
	NSString *value = [[event paramDescriptorForKeyword:keyDirectObject] stringValue];
	if (value != nil) {
		incomingURL((char *)[value UTF8String]);
	}
}
@end

void watchIncomingURLs(void) {
	static FyneIncomingURLHandler *handler;
	if (handler != nil) {
		return;
	}
	handler = [FyneIncomingURLHandler new];
	[[NSAppleEventManager sharedAppleEventManager]
		setEventHandler:handler
		andSelector:@selector(handleURL:withReplyEvent:)
		forEventClass:kInternetEventClass
		andEventID:kAEGetURL];
}
