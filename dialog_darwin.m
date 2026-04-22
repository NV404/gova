// Native macOS dialog bridge. Uses NSAlert for alert/confirm and
// NSOpenPanel / NSSavePanel for file pickers. All presentations dispatch to
// the main queue; completion handlers invoke Go-side callbacks by ID.
//
// Sheet-modal when a key or main window is available; otherwise falls back
// to -runModal so background apps still show a dialog.

#import <AppKit/AppKit.h>

#import "dialog_darwin.h"

static NSWindow* GovaPickParentWindow(void) {
    NSWindow* w = [NSApp keyWindow];
    if (w) return w;
    w = [NSApp mainWindow];
    if (w) return w;
    for (NSWindow* candidate in [NSApp windows]) {
        if ([candidate isVisible]) return candidate;
    }
    return nil;
}

static NSString* GovaStr(const char* s) {
    if (!s) return @"";
    return [NSString stringWithUTF8String:s];
}

static NSArray<NSString*>* GovaExts(const char** extensions, int count) {
    if (!extensions || count <= 0) return nil;
    NSMutableArray<NSString*>* arr = [NSMutableArray arrayWithCapacity:count];
    for (int i = 0; i < count; i++) {
        const char* e = extensions[i];
        if (!e) continue;
        NSString* s = [NSString stringWithUTF8String:e];
        // NSOpenPanel's allowedFileTypes wants bare extensions, no leading dot.
        if ([s hasPrefix:@"."]) s = [s substringFromIndex:1];
        if (s.length > 0) [arr addObject:s];
    }
    return arr.count > 0 ? arr : nil;
}

void govaShowAlert(const char* title, const char* message, const char* ok, unsigned long long cbid) {
    NSString* nsTitle = GovaStr(title);
    NSString* nsMsg = GovaStr(message);
    NSString* nsOK = (ok && ok[0]) ? GovaStr(ok) : @"OK";

    dispatch_async(dispatch_get_main_queue(), ^{
        NSAlert* alert = [[NSAlert alloc] init];
        alert.messageText = nsTitle;
        alert.informativeText = nsMsg;
        alert.alertStyle = NSAlertStyleInformational;
        [alert addButtonWithTitle:nsOK];

        NSWindow* parent = GovaPickParentWindow();
        if (parent) {
            [alert beginSheetModalForWindow:parent completionHandler:^(NSModalResponse r) {
                (void)r;
                govaDialogAlertDone(cbid);
            }];
        } else {
            [alert runModal];
            govaDialogAlertDone(cbid);
        }
    });
}

void govaShowConfirm(
    const char* title,
    const char* message,
    const char* confirmLabel,
    const char* cancelLabel,
    int destructive,
    unsigned long long cbid)
{
    NSString* nsTitle = GovaStr(title);
    NSString* nsMsg = GovaStr(message);
    NSString* nsConfirm = (confirmLabel && confirmLabel[0]) ? GovaStr(confirmLabel) : @"Confirm";
    NSString* nsCancel = (cancelLabel && cancelLabel[0]) ? GovaStr(cancelLabel) : @"Cancel";

    dispatch_async(dispatch_get_main_queue(), ^{
        NSAlert* alert = [[NSAlert alloc] init];
        alert.messageText = nsTitle;
        alert.informativeText = nsMsg;
        alert.alertStyle = destructive ? NSAlertStyleCritical : NSAlertStyleWarning;
        NSButton* confirmBtn = [alert addButtonWithTitle:nsConfirm];
        [alert addButtonWithTitle:nsCancel];
        if (destructive) {
            if (@available(macOS 11.0, *)) {
                confirmBtn.hasDestructiveAction = YES;
            }
        }

        void (^done)(NSModalResponse) = ^(NSModalResponse r) {
            int confirmed = (r == NSAlertFirstButtonReturn) ? 1 : 0;
            govaDialogConfirmDone(cbid, confirmed);
        };

        NSWindow* parent = GovaPickParentWindow();
        if (parent) {
            [alert beginSheetModalForWindow:parent completionHandler:done];
        } else {
            done([alert runModal]);
        }
    });
}

static void GovaApplyStartDir(NSSavePanel* panel, const char* startDir) {
    if (!startDir || !startDir[0]) return;
    NSString* s = [NSString stringWithUTF8String:startDir];
    NSURL* u = [NSURL fileURLWithPath:s isDirectory:YES];
    if (u) panel.directoryURL = u;
}

void govaShowFileOpen(
    const char* startDir,
    const char** extensions,
    int extCount,
    unsigned long long cbid)
{
    NSString* nsStart = startDir ? [NSString stringWithUTF8String:startDir] : nil;
    NSArray<NSString*>* exts = GovaExts(extensions, extCount);

    dispatch_async(dispatch_get_main_queue(), ^{
        NSOpenPanel* panel = [NSOpenPanel openPanel];
        panel.canChooseFiles = YES;
        panel.canChooseDirectories = NO;
        panel.allowsMultipleSelection = NO;
        if (exts) {
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"
            panel.allowedFileTypes = exts;
#pragma clang diagnostic pop
        }
        if (nsStart.length > 0) {
            GovaApplyStartDir(panel, nsStart.UTF8String);
        }

        void (^done)(NSModalResponse) = ^(NSModalResponse r) {
            if (r != NSModalResponseOK || panel.URL == nil) {
                govaDialogPathDone(cbid, (char*)NULL, 1);
                return;
            }
            const char* path = panel.URL.path.UTF8String;
            govaDialogPathDone(cbid, (char*)(path ? path : ""), 0);
        };

        NSWindow* parent = GovaPickParentWindow();
        if (parent) {
            [panel beginSheetModalForWindow:parent completionHandler:done];
        } else {
            done([panel runModal]);
        }
    });
}

void govaShowFileSave(
    const char* startDir,
    const char* defaultName,
    const char** extensions,
    int extCount,
    unsigned long long cbid)
{
    NSString* nsStart = startDir ? [NSString stringWithUTF8String:startDir] : nil;
    NSString* nsName = (defaultName && defaultName[0]) ? [NSString stringWithUTF8String:defaultName] : nil;
    NSArray<NSString*>* exts = GovaExts(extensions, extCount);

    dispatch_async(dispatch_get_main_queue(), ^{
        NSSavePanel* panel = [NSSavePanel savePanel];
        if (nsName.length > 0) {
            panel.nameFieldStringValue = nsName;
        }
        if (exts) {
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"
            panel.allowedFileTypes = exts;
#pragma clang diagnostic pop
        }
        if (nsStart.length > 0) {
            GovaApplyStartDir(panel, nsStart.UTF8String);
        }

        void (^done)(NSModalResponse) = ^(NSModalResponse r) {
            if (r != NSModalResponseOK || panel.URL == nil) {
                govaDialogPathDone(cbid, (char*)NULL, 1);
                return;
            }
            const char* path = panel.URL.path.UTF8String;
            govaDialogPathDone(cbid, (char*)(path ? path : ""), 0);
        };

        NSWindow* parent = GovaPickParentWindow();
        if (parent) {
            [panel beginSheetModalForWindow:parent completionHandler:done];
        } else {
            done([panel runModal]);
        }
    });
}

void govaShowFolderOpen(const char* startDir, unsigned long long cbid) {
    NSString* nsStart = startDir ? [NSString stringWithUTF8String:startDir] : nil;

    dispatch_async(dispatch_get_main_queue(), ^{
        NSOpenPanel* panel = [NSOpenPanel openPanel];
        panel.canChooseFiles = NO;
        panel.canChooseDirectories = YES;
        panel.allowsMultipleSelection = NO;
        panel.canCreateDirectories = YES;
        if (nsStart.length > 0) {
            GovaApplyStartDir(panel, nsStart.UTF8String);
        }

        void (^done)(NSModalResponse) = ^(NSModalResponse r) {
            if (r != NSModalResponseOK || panel.URL == nil) {
                govaDialogPathDone(cbid, (char*)NULL, 1);
                return;
            }
            const char* path = panel.URL.path.UTF8String;
            govaDialogPathDone(cbid, (char*)(path ? path : ""), 0);
        };

        NSWindow* parent = GovaPickParentWindow();
        if (parent) {
            [panel beginSheetModalForWindow:parent completionHandler:done];
        } else {
            done([panel runModal]);
        }
    });
}
