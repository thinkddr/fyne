// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ios

#include "_cgo_export.h"
#include <pthread.h>
#include <stdio.h>
#include <sys/utsname.h>
#include "ios_view_geometry.h"

#import <UIKit/UIKit.h>
#import <MobileCoreServices/MobileCoreServices.h>
#import <GLKit/GLKit.h>
#import <UserNotifications/UserNotifications.h>
#import <BackgroundTasks/BackgroundTasks.h>

struct utsname sysInfo;

@interface GoAppAppController : GLKViewController<UIContentContainer, GLKViewDelegate>
- (void)updateGeometry;
@end

@interface GoInputView : UITextField<UITextFieldDelegate>
@end

@interface GoAppAppDelegate : UIResponder<UIApplicationDelegate>
@property (strong, nonatomic) UIWindow *window;
@property (strong, nonatomic) GoAppAppController *controller;
@end

@interface GoAppSceneDelegate : UIResponder<UIWindowSceneDelegate>
@property (strong, nonatomic) UIWindow *window;
// The options of a scene that connected without being in the foreground: what waited
// for a user interface, and the scene's URLs after it, wait for its first foreground.
@property (strong, nonatomic) UISceneConnectionOptions *pendingConnection;
@end

// The app refresh task is the first entry of BGTaskSchedulerPermittedIdentifiers; an
// app that declares none never gets one. See mobile.BackgroundDriver.
static NSString *backgroundRefreshIdentifier(void) {
	NSArray *identifiers = [[NSBundle mainBundle] objectForInfoDictionaryKey:@"BGTaskSchedulerPermittedIdentifiers"];
	if (![identifiers isKindOfClass:[NSArray class]] || identifiers.count == 0 || ![identifiers[0] isKindOfClass:[NSString class]]) {
		return nil;
	}
	return identifiers[0];
}

// Must run before launching finishes: BGTaskScheduler throws afterwards.
static void registerBackgroundRefresh(void) {
	NSString *identifier = backgroundRefreshIdentifier();
	if (identifier == nil) {
		return;
	}
	[[BGTaskScheduler sharedScheduler] registerForTaskWithIdentifier:identifier usingQueue:nil launchHandler:^(BGTask *task) {
		// Retained until completeBackgroundRefresh, which Go calls exactly once.
		uintptr_t handle = (uintptr_t)CFBridgingRetain(task);
		// Known to Go before it can expire: a task that expires before its handler starts
		// (the main thread busy) is completed at once and the handler never runs.
		backgroundRefreshReceived(handle);
		task.expirationHandler = ^{
			backgroundRefreshExpired(handle);
		};
		// The launch handler runs on a background queue, and the delegate's protected
		// data notifications can reach a suspended app late: ask UIKit now, on the main
		// thread, before the handler reads anything.
		dispatch_async(dispatch_get_main_queue(), ^{
			protectedDataChanged([UIApplication sharedApplication].isProtectedDataAvailable);
			backgroundRefreshStarted(handle);
		});
	}];
}

// A new request replaces the pending one. Zero or less requests nothing.
static void scheduleBackgroundRefresh(double seconds) {
	NSString *identifier = backgroundRefreshIdentifier();
	if (seconds <= 0 || identifier == nil) {
		return;
	}
	BGAppRefreshTaskRequest *request = [[BGAppRefreshTaskRequest alloc] initWithIdentifier:identifier];
	request.earliestBeginDate = [NSDate dateWithTimeIntervalSinceNow:seconds];
	NSError *error = nil;
	if (![[BGTaskScheduler sharedScheduler] submitTaskRequest:request error:&error]) {
		NSLog(@"fyne: the app refresh task was not scheduled: %@", error);
	}
}

void completeBackgroundRefresh(uintptr_t handle, bool ok, double next) {
	dispatch_async(dispatch_get_main_queue(), ^{
		BGTask *task = (BGTask *)CFBridgingRelease((CFTypeRef)handle);
		scheduleBackgroundRefresh(next);
		[task setTaskCompleted:ok ? YES : NO];
	});
}

@implementation GoAppAppDelegate
- (BOOL)application:(UIApplication *)application didFinishLaunchingWithOptions:(NSDictionary *)launchOptions {
    int scale = 1;
    if ([[UIScreen mainScreen] respondsToSelector:@selector(displayLinkWithTarget:selector:)]) {
		scale = (int)[UIScreen mainScreen].scale; // either 1.0, 2.0, or 3.0.
	}
    CGSize size = [UIScreen mainScreen].nativeBounds.size;
    setDisplayMetrics((int)size.width, (int)size.height, scale);

	lifecycleAlive();
	// The window is created by GoAppSceneDelegate: from the iOS 27 SDK on, UIKit refuses
	// to launch an app that does not adopt the scene lifecycle (it traps in
	// _UIApplicationEvaluateRuntimeIssueForNoSceneLifecycleAdoption).

	UNUserNotificationCenter *center = [UNUserNotificationCenter currentNotificationCenter];
	center.delegate = (id) self;

	NSURL *incoming = launchOptions[UIApplicationLaunchOptionsURLKey];
	if (incoming != nil) {
		urlOpened((char *)[[incoming absoluteString] UTF8String]);
	}

	registerBackgroundRefresh();
	// Without a scene yet: UIKit launched the app for background work if the state is
	// background now. Protected data is unavailable when it launched with the device locked.
	launchFinished(application.applicationState == UIApplicationStateBackground, application.isProtectedDataAvailable);
	return YES;
}

- (void)applicationProtectedDataWillBecomeUnavailable:(UIApplication *)application {
	protectedDataChanged(false);
}

- (void)applicationProtectedDataDidBecomeAvailable:(UIApplication *)application {
	protectedDataChanged(true);
}

- (UISceneConfiguration *)application:(UIApplication *)application configurationForConnectingSceneSession:(UISceneSession *)session options:(UISceneConnectionOptions *)options {
	UISceneConfiguration *config = [[UISceneConfiguration alloc] initWithName:@"Default" sessionRole:session.role];
	config.delegateClass = [GoAppSceneDelegate class];
	return config;
}

- (BOOL)application:(UIApplication *)application openURL:(NSURL *)url options:(NSDictionary<UIApplicationOpenURLOptionsKey,id> *)options {
	if (url != nil) {
		urlOpened((char *)[[url absoluteString] UTF8String]);
	}
	return YES;
}

// Universal Links: an https link the app is associated with (apple-app-site-association
// plus the associated-domains entitlement) arrives as a browsing-web user activity, not
// through openURL. iOS calls this both on a cold start (after didFinishLaunching) and
// while running, so the URL goes to the same handler as a custom scheme.
- (BOOL)application:(UIApplication *)application continueUserActivity:(NSUserActivity *)userActivity restorationHandler:(void (^)(NSArray<id<UIUserActivityRestoring>> *))restorationHandler {
	if ([userActivity.activityType isEqualToString:NSUserActivityTypeBrowsingWeb] && userActivity.webpageURL != nil) {
		urlOpened((char *)[[userActivity.webpageURL absoluteString] UTF8String]);
		return YES;
	}
	return NO;
}

- (void)applicationDidBecomeActive:(UIApplication * )application {
	lifecycleFocused();
}

- (void)applicationWillResignActive:(UIApplication *)application {
	lifecycleVisible();
}

- (void)applicationDidEnterBackground:(UIApplication *)application {
	lifecycleAlive();
	scheduleBackgroundRefresh(backgroundRefreshInterval());
}

- (void)applicationWillTerminate:(UIApplication *)application {
	lifecycleDead();
}

- (void)applicationDidReceiveMemoryWarning:(UIApplication *)application {
	lifecycleMemoryWarning();
}

- (void)documentPicker:(UIDocumentPickerViewController *)controller didPickDocumentsAtURLs:(NSArray <NSURL *>*)urls {
    if ([urls count] == 0) {
        return;
    }

    NSURL* url = urls[0];
    NSURL* toClose = NULL;
    BOOL secured = [url startAccessingSecurityScopedResource];
    if (secured) {
        toClose = url;
    }

    filePickerReturned((char*)[[url description] UTF8String], toClose);
}

- (void)documentPickerWasCancelled:(UIDocumentPickerViewController *)controller {
    filePickerReturned("", NULL);
}

- (void)userNotificationCenter:(UNUserNotificationCenter *)center
       willPresentNotification:(UNNotification *)notification
         withCompletionHandler:(void (^)(UNNotificationPresentationOptions options))completionHandler {
	completionHandler(UNNotificationPresentationOptionAlert);
}
@end

// GoAppSceneDelegate owns the window and receives what the scene lifecycle moves away
// from the application delegate: incoming URLs, Universal Links and the foreground and
// background transitions.
@implementation GoAppSceneDelegate
- (void)scene:(UIScene *)scene willConnectToSession:(UISceneSession *)session options:(UISceneConnectionOptions *)connectionOptions {
	if (![scene isKindOfClass:[UIWindowScene class]]) {
		return;
	}
	GoAppAppDelegate *app = (GoAppAppDelegate *)[UIApplication sharedApplication].delegate;
	self.window = [[UIWindow alloc] initWithWindowScene:(UIWindowScene *)scene];
	app.controller = [[GoAppAppController alloc] initWithNibName:nil bundle:nil];
	self.window.rootViewController = app.controller;
	app.window = self.window;
	[self.window makeKeyAndVisible];

	// update insets once key window is set
	[app.controller updateGeometry];

	// UIKit can connect a scene that is not in the foreground (a system request during
	// background work): nobody is looking, so what waits for a user interface waits
	// for the scene's first foreground, and its URLs after it.
	UISceneActivationState state = scene.activationState;
	if (state != UISceneActivationStateForegroundActive && state != UISceneActivationStateForegroundInactive) {
		self.pendingConnection = connectionOptions;
		return;
	}
	[self scene:scene connectWithOptions:connectionOptions];
}

- (void)scene:(UIScene *)scene connectWithOptions:(UISceneConnectionOptions *)connectionOptions {
	// What the app deferred until it had a user interface, before this scene's URLs.
	sceneWillConnect();
	[self scene:scene openURLContexts:connectionOptions.URLContexts];
	for (NSUserActivity *activity in connectionOptions.userActivities) {
		[self scene:scene continueUserActivity:activity];
	}
}

- (void)sceneWillEnterForeground:(UIScene *)scene {
	UISceneConnectionOptions *pending = self.pendingConnection;
	if (pending != nil) {
		self.pendingConnection = nil;
		[self scene:scene connectWithOptions:pending];
	}
}

- (void)scene:(UIScene *)scene openURLContexts:(NSSet<UIOpenURLContext *> *)contexts {
	for (UIOpenURLContext *c in contexts) {
		if (c.URL != nil) {
			urlOpened((char *)[[c.URL absoluteString] UTF8String]);
		}
	}
}

- (void)scene:(UIScene *)scene continueUserActivity:(NSUserActivity *)userActivity {
	if ([userActivity.activityType isEqualToString:NSUserActivityTypeBrowsingWeb] && userActivity.webpageURL != nil) {
		urlOpened((char *)[[userActivity.webpageURL absoluteString] UTF8String]);
	}
}

- (void)sceneDidBecomeActive:(UIScene *)scene {
	lifecycleFocused();
}

- (void)sceneWillResignActive:(UIScene *)scene {
	lifecycleVisible();
}

- (void)sceneDidEnterBackground:(UIScene *)scene {
	lifecycleAlive();
	scheduleBackgroundRefresh(backgroundRefreshInterval());
}
@end

@interface GoAppAppController ()
@property (strong, nonatomic) EAGLContext *context;
@property (strong, nonatomic) GLKView *glview;
@property (strong, nonatomic) GoInputView *inputView;
@property (nonatomic) CGRect keyboardFrame;
@end

@implementation GoAppAppController
- (void)viewWillAppear:(BOOL)animated
{
	// TODO: replace by swapping out GLKViewController for a UIVIewController.
	[super viewWillAppear:animated];
	self.paused = YES;

    [[NSNotificationCenter defaultCenter] addObserver:self selector:@selector(keyboardWillChangeFrame:) name:UIKeyboardWillChangeFrameNotification object:nil];
    [[NSNotificationCenter defaultCenter] addObserver:self selector:@selector(keyboardWillHide:) name:UIKeyboardWillHideNotification object:nil];
}

- (void)viewWillDisappear:(BOOL)animated {
    [super viewWillDisappear:animated];

    [[NSNotificationCenter defaultCenter] removeObserver:self name:UIKeyboardWillChangeFrameNotification object:nil];
    [[NSNotificationCenter defaultCenter] removeObserver:self name:UIKeyboardWillHideNotification object:nil];
}

- (void)viewDidLoad {
	[super viewDidLoad];
	self.context = [[EAGLContext alloc] initWithAPI:kEAGLRenderingAPIOpenGLES2];
	self.inputView = [[GoInputView alloc] initWithFrame:CGRectMake(0, 0, 0, 0)];
	self.inputView.delegate = self.inputView;
	self.inputView.autocapitalizationType = UITextAutocapitalizationTypeNone;
	self.inputView.autocorrectionType = UITextAutocorrectionTypeNo;
	[self.view addSubview:self.inputView];
	self.glview = (GLKView*)self.view;
	self.glview.drawableDepthFormat = GLKViewDrawableDepthFormat24;
	self.glview.multipleTouchEnabled = true; // TODO expose setting to user.
	self.glview.context = self.context;
	self.glview.userInteractionEnabled = YES;
	self.keyboardFrame = CGRectNull;
	//self.glview.enableSetNeedsDisplay = YES; // only invoked once

	// Do not use the GLKViewController draw loop.
	//self.paused = YES;
	//self.resumeOnDidBecomeActive = NO;
	//self.preferredFramesPerSecond = 0;

    self.glview.enableSetNeedsDisplay = NO;
    CADisplayLink* displayLink = [CADisplayLink displayLinkWithTarget:self selector:@selector(render:)];
    [displayLink addToRunLoop:[NSRunLoop currentRunLoop] forMode:NSDefaultRunLoopMode];
}

- (void)updateGeometry {
	UIWindow *window = self.view.window;
	UIWindowScene *scene = window.windowScene;
	if (scene == nil) {
		return; // viewDidLoad can run before the view has joined its scene.
	}
	CGRect bounds = self.view.bounds;
	CGFloat scale = self.view.contentScaleFactor;
	int width = ios_view_pixels(bounds.size.width, scale);
	int height = ios_view_pixels(bounds.size.height, scale);
	if (width == 0 || height == 0) {
		return;
	}
	UIInterfaceOrientation orientation;
	if (@available(iOS 16.0, *)) {
		orientation = scene.effectiveGeometry.interfaceOrientation;
	} else {
		orientation = scene.interfaceOrientation;
	}
	UIEdgeInsets inset = self.view.safeAreaInsets;
	if (!CGRectIsNull(self.keyboardFrame)) {
		CGRect keyboard = [self.view convertRect:self.keyboardFrame fromCoordinateSpace:window.screen.coordinateSpace];
		CGFloat bottom = ios_keyboard_bottom_inset(bounds.size.width, bounds.size.height,
			CGRectGetMinX(keyboard) - CGRectGetMinX(bounds), CGRectGetMinY(keyboard) - CGRectGetMinY(bounds),
			keyboard.size.width, keyboard.size.height);
		inset.bottom = MAX(inset.bottom, bottom);
	}
	setScreen((float)scale);
	updateConfig(width, height, orientation,
		ios_view_pixels(inset.top, scale), ios_view_pixels(inset.bottom, scale),
		ios_view_pixels(inset.left, scale), ios_view_pixels(inset.right, scale));
}

- (void)viewDidLayoutSubviews {
	[super viewDidLayoutSubviews];
	[self updateGeometry];
}

- (void)viewSafeAreaInsetsDidChange {
	[super viewSafeAreaInsetsDidChange];
	[self updateGeometry];
}

- (void)viewWillTransitionToSize:(CGSize)ptSize withTransitionCoordinator:(id<UIViewControllerTransitionCoordinator>)coordinator {
	[super viewWillTransitionToSize:ptSize withTransitionCoordinator:coordinator];
	[coordinator animateAlongsideTransition:^(id<UIViewControllerTransitionCoordinatorContext> context) {
		// TODO(crawshaw): come up with a plan to handle animations.
	} completion:^(id<UIViewControllerTransitionCoordinatorContext> context) {
		[self updateGeometry];
	}];
}

- (void)render:(CADisplayLink*)displayLink {
    if (needsDraw()) {
        [self.glview display];
    }
}

- (void)glkView:(GLKView *)view drawInRect:(CGRect)rect {
    drawloop();
}

#define TOUCH_TYPE_BEGIN 0 // touch.TypeBegin
#define TOUCH_TYPE_MOVE  1 // touch.TypeMove
#define TOUCH_TYPE_END   2 // touch.TypeEnd
#define TOUCH_TYPE_CANCEL 3 // touch.TypeCancel

static void sendTouches(int change, NSSet* touches, UIView *view) {
	CGFloat scale = view.contentScaleFactor;
	for (UITouch* touch in touches) {
		CGPoint p = [touch locationInView:view];
		sendTouch((GoUintptr)touch, (GoUintptr)change, p.x*scale, p.y*scale);
	}
}

- (void)touchesBegan:(NSSet*)touches withEvent:(UIEvent*)event {
	sendTouches(TOUCH_TYPE_BEGIN, touches, self.view);
}

- (void)touchesMoved:(NSSet*)touches withEvent:(UIEvent*)event {
	sendTouches(TOUCH_TYPE_MOVE, touches, self.view);
}

- (void)touchesEnded:(NSSet*)touches withEvent:(UIEvent*)event {
	sendTouches(TOUCH_TYPE_END, touches, self.view);
}

- (void)touchesCancelled:(NSSet*)touches withEvent:(UIEvent*)event {
    sendTouches(TOUCH_TYPE_CANCEL, touches, self.view);
}

- (void) traitCollectionDidChange: (UITraitCollection *) previousTraitCollection {
    [super traitCollectionDidChange: previousTraitCollection];

	[self updateGeometry];
}

- (void)keyboardWillChangeFrame:(NSNotification *)note {
	NSValue *frame = note.userInfo[UIKeyboardFrameEndUserInfoKey];
	self.keyboardFrame = frame == nil ? CGRectNull : frame.CGRectValue;
	[self updateGeometry];
}

- (void)keyboardWillHide:(NSNotification *)note {
	self.keyboardFrame = CGRectNull;
	[self updateGeometry];
}

@end

@implementation GoInputView

- (BOOL)canBecomeFirstResponder {
    return YES;
}

- (void)deleteBackward {
    keyboardDelete();
}

-(BOOL)textField:(UITextField *)textField shouldChangeCharactersInRange:(NSRange)range replacementString:(NSString *)string {
    keyboardTyped((char *)[string UTF8String]);
    return NO;
}

- (BOOL)textFieldShouldReturn:(UITextField *)textField {
    if ([self returnKeyType] != UIReturnKeyDone) {
        keyboardTyped("\n");
        return YES;
    }

    dispatch_async(dispatch_get_main_queue(), ^{
        [self resignFirstResponder];
    });

    return NO;
}

@end

void runApp(void) {
	char * argv[] = {};
	@autoreleasepool {
		UIApplicationMain(0, argv, nil, NSStringFromClass([GoAppAppDelegate class]));
	}
}

void makeCurrentContext(GLintptr context) {
	EAGLContext* ctx = (EAGLContext*)context;
	if (![EAGLContext setCurrentContext:ctx]) {
		// TODO(crawshaw): determine how terrible this is. Exit?
		NSLog(@"failed to set current context");
	}
}

void swapBuffers(GLintptr context) {
	__block EAGLContext* ctx = (EAGLContext*)context;
	dispatch_sync(dispatch_get_main_queue(), ^{
		[EAGLContext setCurrentContext:ctx];
		[ctx presentRenderbuffer:GL_RENDERBUFFER];
	});
}

uint64_t threadID() {
	uint64_t id;
	if (pthread_threadid_np(pthread_self(), &id)) {
		abort();
	}
	return id;
}

bool isDark() {
    UIViewController *rootVC = [[[[UIApplication sharedApplication] delegate] window] rootViewController];
    return rootVC.traitCollection.userInterfaceStyle == UIUserInterfaceStyleDark;
}

#define DEFAULT_KEYBOARD_CODE 0
#define SINGLELINE_KEYBOARD_CODE 1
#define NUMBER_KEYBOARD_CODE 2

void showKeyboard(int keyboardType) {
    GoAppAppDelegate *appDelegate = (GoAppAppDelegate *)[[UIApplication sharedApplication] delegate];
    GoInputView *view = appDelegate.controller.inputView;

    dispatch_async(dispatch_get_main_queue(), ^{
        switch (keyboardType)
        {
            case DEFAULT_KEYBOARD_CODE:
                [view setKeyboardType:UIKeyboardTypeDefault];
                [view setReturnKeyType:UIReturnKeyDefault];
                break;
            case SINGLELINE_KEYBOARD_CODE:
                [view setKeyboardType:UIKeyboardTypeDefault];
                [view setReturnKeyType:UIReturnKeyDone];
                break;
            case NUMBER_KEYBOARD_CODE:
                [view setKeyboardType:UIKeyboardTypeNumberPad];
                [view setReturnKeyType:UIReturnKeyDone];
                break;
            default:
                NSLog(@"unknown keyboard type, use default");
                [view setKeyboardType:UIKeyboardTypeDefault];
                [view setReturnKeyType:UIReturnKeyDefault];
                break;
        }
        // refresh settings if keyboard is already open
        [view reloadInputViews];

        BOOL ret = [view becomeFirstResponder];
    });
}

void hideKeyboard() {
    GoAppAppDelegate *appDelegate = (GoAppAppDelegate *)[[UIApplication sharedApplication] delegate];
    GoInputView *view = appDelegate.controller.inputView;

    dispatch_async(dispatch_get_main_queue(), ^{
        [view resignFirstResponder];
    });
}

NSMutableArray *docTypesForMimeExts(char *mimes, char *exts) {
    NSMutableArray *docTypes = [NSMutableArray array];
    if (mimes != NULL && strlen(mimes) > 0) {
        NSString *mimeList = [NSString stringWithUTF8String:mimes];

        if ([mimeList isEqualToString:@"application/x-directory"]) {
            [docTypes addObject:(NSString*)kUTTypeFolder];
        } else {
            NSArray *mimeItems = [mimeList componentsSeparatedByString:@"|"];

            for (NSString *mime in mimeItems)  {
                NSString *UTI = (NSString *) UTTypeCreatePreferredIdentifierForTag(kUTTagClassMIMEType, (CFStringRef)mime, NULL);

                [docTypes addObject:UTI];
            }
        }
    } else if (exts != NULL && strlen(exts) > 0) {
        NSString *extList = [NSString stringWithUTF8String:exts];
        NSArray *extItems = [extList componentsSeparatedByString:@"|"];

        for (NSString *ext in extItems)  {
            NSString *UTI = (NSString *) UTTypeCreatePreferredIdentifierForTag(kUTTagClassFilenameExtension, (CFStringRef)ext, NULL);

            [docTypes addObject:UTI];
        }
    } else {
        [docTypes addObject:@"public.data"];
    }

    return docTypes;
}

void showFileOpenPicker(char* mimes, char *exts) {
    GoAppAppDelegate *appDelegate = (GoAppAppDelegate *)[[UIApplication sharedApplication] delegate];

    NSMutableArray *docTypes = docTypesForMimeExts(mimes, exts);

    UIDocumentPickerViewController *documentPicker = [[UIDocumentPickerViewController alloc]
        initWithDocumentTypes:docTypes inMode:UIDocumentPickerModeOpen];
    documentPicker.delegate = (id) appDelegate;

    dispatch_async(dispatch_get_main_queue(), ^{
        [appDelegate.controller presentViewController:documentPicker animated:YES completion:nil];
    });
}

void showFileSavePicker(char* mimes, char *exts) {
    GoAppAppDelegate *appDelegate = (GoAppAppDelegate *)[[UIApplication sharedApplication] delegate];

    NSMutableArray *docTypes = docTypesForMimeExts(mimes, exts);

    NSURL *temporaryDirectoryURL = [NSURL fileURLWithPath: NSTemporaryDirectory() isDirectory: YES];
    NSURL *temporaryFileURL = [temporaryDirectoryURL URLByAppendingPathComponent:@"filename"];

    char* bytes = "\n";
    NSData *data = [NSData dataWithBytes:bytes length:1];
    BOOL ok = [data writeToURL:temporaryFileURL atomically:YES];

    UIDocumentPickerViewController *documentPicker = [[UIDocumentPickerViewController alloc]
        initWithURL:temporaryFileURL inMode:UIDocumentPickerModeMoveToService];
    documentPicker.delegate = (id) appDelegate;

    dispatch_async(dispatch_get_main_queue(), ^{
        [appDelegate.controller presentViewController:documentPicker animated:YES completion:nil];
    });
}

void closeFileResource(void* urlPtr) {
    if (urlPtr == NULL) {
        return;
    }

    NSURL* url = (NSURL*) urlPtr;
    [url stopAccessingSecurityScopedResource];
}
