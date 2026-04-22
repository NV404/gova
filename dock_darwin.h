// Declarations for the macOS dock helpers. The implementations live in
// dock_darwin.m so the Objective-C classes are compiled exactly once; cgo
// pulls in the preamble twice (once for the Go file, once for the //export
// bridge C file) which would otherwise produce duplicate symbols.

#ifndef GOVA_DOCK_DARWIN_H
#define GOVA_DOCK_DARWIN_H

#ifdef __cplusplus
extern "C" {
#endif

void govaDockSetBadge(const char* text);
void govaDockBounce(void);
void govaDockSetProgress(double frac);
void govaDockSetMenu(const char** labels, const unsigned long long* ids, int count);
void govaSetAppIcon(const void* data, int length);

extern void govaDockMenuFire(unsigned long long id);

#ifdef __cplusplus
}
#endif

#endif
