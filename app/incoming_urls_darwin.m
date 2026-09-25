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

// Finder, `open file.ext` and "Open With" send kAEOpenDocuments with a list of files,
// not argv. Each one is delivered as a file:// URL through the same handler.
- (void)handleOpenDocuments:(NSAppleEventDescriptor *)event withReplyEvent:(NSAppleEventDescriptor *)reply {
	NSAppleEventDescriptor *list = [event paramDescriptorForKeyword:keyDirectObject];
	NSInteger count = [list numberOfItems];
	for (NSInteger i = count == 0 ? 0 : 1; i <= count; i++) {
		NSAppleEventDescriptor *item = count == 0 ? list : [list descriptorAtIndex:i];
		NSAppleEventDescriptor *file = [item coerceToDescriptorType:typeFileURL];
		if (file == nil) {
			continue;
		}
		NSString *value = [[[NSString alloc] initWithData:[file data] encoding:NSUTF8StringEncoding] autorelease];
		if (value != nil) {
			incomingURL((char *)[value UTF8String]);
		}
	}
}

- (void)installOpenDocuments {
	[[NSAppleEventManager sharedAppleEventManager]
		setEventHandler:self
		andSelector:@selector(handleOpenDocuments:withReplyEvent:)
		forEventClass:kCoreEventClass
		andEventID:kAEOpenDocuments];
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
	// AppKit installs its own kAEOpenDocuments handler while finishing launching, which
	// would replace one installed now; the documents of the launch arrive right after
	// applicationWillFinishLaunching, so that is where ours goes. Installed now too, for
	// an application that has already finished launching.
	[handler installOpenDocuments];
	[[NSNotificationCenter defaultCenter]
		addObserver:handler
		selector:@selector(installOpenDocuments)
		name:NSApplicationWillFinishLaunchingNotification
		object:nil];
}
