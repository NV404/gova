// Objective-C runtime for macOS dock integration.
//
// This file is compiled once by cgo as an auxiliary source. Keeping the
// @interface/@implementation pairs out of the cgo preamble prevents
// duplicate-symbol link errors caused by the preamble being emitted into
// both the Go file's .o and the _cgo_export.c helper's .o.

#import <AppKit/AppKit.h>
#import <objc/runtime.h>

#import "dock_darwin.h"

void govaDockSetBadge(const char* text) {
    NSString* s = (text && text[0]) ? [NSString stringWithUTF8String:text] : @"";
    dispatch_async(dispatch_get_main_queue(), ^{
        NSDockTile* tile = [NSApp dockTile];
        [tile setBadgeLabel:s];
        [tile display];
    });
}

void govaDockBounce(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        [NSApp requestUserAttention:NSInformationalRequest];
    });
}

// Sets the app icon shown in the dock and Cmd-Tab switcher for the running
// process. Fyne's SetIcon only stores the resource; it never calls AppKit,
// so for an unbundled `go run` binary the dock stays on the default Go
// terminal icon unless we do this ourselves.
void govaSetAppIcon(const void* data, int length) {
    if (!data || length <= 0) return;
    NSData* payload = [NSData dataWithBytes:data length:length];
    dispatch_async(dispatch_get_main_queue(), ^{
        NSImage* image = [[NSImage alloc] initWithData:payload];
        if (!image) return;
        [NSApp setApplicationIconImage:image];
    });
}

@interface GovaDockProgressView : NSView
@property(nonatomic) double progress;
@end

@implementation GovaDockProgressView
- (void)drawRect:(NSRect)dirty {
    NSRect b = self.bounds;
    NSImage* icon = [NSApp applicationIconImage];
    if (icon) {
        [icon drawInRect:b
                fromRect:NSZeroRect
               operation:NSCompositingOperationCopy
                fraction:1.0
          respectFlipped:YES
                   hints:nil];
    }
    double p = self.progress;
    if (p < 0) p = 0;
    if (p > 1) p = 1;

    CGFloat padX = b.size.width * 0.08;
    CGFloat h = MAX(10.0, b.size.height * 0.10);
    CGFloat y = b.size.height * 0.08;
    NSRect bg = NSMakeRect(padX, y, b.size.width - padX*2, h);
    NSBezierPath* bgPath = [NSBezierPath bezierPathWithRoundedRect:bg xRadius:h/2 yRadius:h/2];
    [[NSColor colorWithWhite:0.0 alpha:0.55] set];
    [bgPath fill];

    NSRect fg = bg;
    fg.size.width = bg.size.width * p;
    NSBezierPath* fgPath = [NSBezierPath bezierPathWithRoundedRect:fg xRadius:h/2 yRadius:h/2];
    [[NSColor colorWithRed:0.23 green:0.55 blue:0.98 alpha:1.0] set];
    [fgPath fill];
}
@end

static GovaDockProgressView* govaProgressView = nil;

void govaDockSetProgress(double frac) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSDockTile* tile = [NSApp dockTile];
        if (frac < 0) {
            [tile setContentView:nil];
            [tile display];
            return;
        }
        if (!govaProgressView) {
            govaProgressView = [[GovaDockProgressView alloc]
                initWithFrame:NSMakeRect(0, 0, 128, 128)];
        }
        govaProgressView.progress = frac;
        [tile setContentView:govaProgressView];
        [govaProgressView setNeedsDisplay:YES];
        [tile display];
    });
}

@interface GovaDockTarget : NSObject
- (void)fire:(id)sender;
@end

@implementation GovaDockTarget
- (void)fire:(id)sender {
    if (![sender isKindOfClass:[NSMenuItem class]]) return;
    NSMenuItem* item = (NSMenuItem*)sender;
    unsigned long long mid = (unsigned long long)item.tag;
    if (mid != 0) govaDockMenuFire(mid);
}
@end

static GovaDockTarget* govaDockTarget = nil;
static NSMenu* govaDockMenu = nil;

static NSMenu* GovaProvideDockMenu(id self, SEL _cmd, NSApplication* app) {
    return govaDockMenu;
}

static BOOL govaInstalledDockMenuMethod = NO;

static void govaInstallDockMenuHandler(void) {
    if (govaInstalledDockMenuMethod) return;
    id delegate = [NSApp delegate];
    if (!delegate) return;
    Class cls = [delegate class];
    SEL sel = @selector(applicationDockMenu:);
    if ([delegate respondsToSelector:sel]) {
        govaInstalledDockMenuMethod = YES;
        return;
    }
    class_addMethod(cls, sel, (IMP)GovaProvideDockMenu, "@@:@");
    govaInstalledDockMenuMethod = YES;
}

void govaDockSetMenu(const char** labels, const unsigned long long* ids, int count) {
    NSMutableArray<NSString*>* labelArr = [NSMutableArray arrayWithCapacity:count];
    for (int i = 0; i < count; i++) {
        NSString* s = labels[i] ? [NSString stringWithUTF8String:labels[i]] : @"";
        [labelArr addObject:s];
    }
    NSMutableArray<NSNumber*>* idArr = [NSMutableArray arrayWithCapacity:count];
    for (int i = 0; i < count; i++) {
        [idArr addObject:@(ids[i])];
    }
    dispatch_async(dispatch_get_main_queue(), ^{
        if (!govaDockTarget) govaDockTarget = [[GovaDockTarget alloc] init];
        NSMenu* menu = [[NSMenu alloc] initWithTitle:@""];
        menu.autoenablesItems = NO;
        for (int i = 0; i < (int)labelArr.count; i++) {
            NSString* label = labelArr[i];
            unsigned long long mid = [idArr[i] unsignedLongLongValue];
            if (label.length == 0) {
                [menu addItem:[NSMenuItem separatorItem]];
                continue;
            }
            NSMenuItem* item = [[NSMenuItem alloc] initWithTitle:label
                                                          action:@selector(fire:)
                                                   keyEquivalent:@""];
            item.target = govaDockTarget;
            item.tag = (NSInteger)mid;
            item.enabled = (mid != 0);
            [menu addItem:item];
        }
        govaDockMenu = menu;
        govaInstallDockMenuHandler();
    });
}
