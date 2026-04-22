// Declarations for the macOS native dialog bridge. Implementations live in
// dialog_darwin.m so the Objective-C code is compiled exactly once; cgo
// emits the preamble twice (Go file + _cgo_export.c), so any @interface /
// @implementation in the preamble would produce duplicate symbols.

#ifndef GOVA_DIALOG_DARWIN_H
#define GOVA_DIALOG_DARWIN_H

#ifdef __cplusplus
extern "C" {
#endif

// Alert / confirm.
void govaShowAlert(
    const char* title,
    const char* message,
    const char* ok,
    unsigned long long cbid);

void govaShowConfirm(
    const char* title,
    const char* message,
    const char* confirmLabel,
    const char* cancelLabel,
    int destructive,
    unsigned long long cbid);

// File pickers. extensions is a NULL-terminated array of C strings; pass
// NULL or an empty list to allow any file type.
void govaShowFileOpen(
    const char* startDir,
    const char** extensions,
    int extCount,
    unsigned long long cbid);

void govaShowFileSave(
    const char* startDir,
    const char* defaultName,
    const char** extensions,
    int extCount,
    unsigned long long cbid);

void govaShowFolderOpen(
    const char* startDir,
    unsigned long long cbid);

extern void govaDialogAlertDone(unsigned long long cbid);
extern void govaDialogConfirmDone(unsigned long long cbid, int confirmed);
extern void govaDialogPathDone(unsigned long long cbid, char* path, int cancelled);

#ifdef __cplusplus
}
#endif

#endif
