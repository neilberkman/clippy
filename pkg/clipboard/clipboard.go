package clipboard

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit
#import <Foundation/Foundation.h>
#import <AppKit/NSPasteboard.h>
#import <AppKit/NSApplication.h>

// Function to copy a file reference to the clipboard
void copyFile(const char *path) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSURL *fileURL = [NSURL fileURLWithPath:[NSString stringWithUTF8String:path]];
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        [pasteboard clearContents];
        [pasteboard writeObjects:@[fileURL]];
    }
}

// Function to copy multiple file references to the clipboard
void copyFiles(const char **paths, int count) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSMutableArray *fileURLs = [NSMutableArray arrayWithCapacity:count];

        for (int i = 0; i < count; i++) {
            NSURL *fileURL = [NSURL fileURLWithPath:[NSString stringWithUTF8String:paths[i]]];
            [fileURLs addObject:fileURL];
        }

        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        [pasteboard clearContents];
        [pasteboard writeObjects:fileURLs];
    }
}

// Function to copy plain text content to the clipboard
void copyText(const char *text) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSString *nsText = [NSString stringWithUTF8String:text];
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        [pasteboard clearContents];
        [pasteboard setString:nsText forType:NSPasteboardTypeString];
    }
}

// Get current clipboard file paths if any
char** getClipboardFiles(int *count) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        NSArray *files = [pasteboard readObjectsForClasses:@[[NSURL class]]
                                                   options:@{NSPasteboardURLReadingFileURLsOnlyKey: @YES}];

        *count = (int)[files count];
        if (*count == 0) return NULL;

        char **paths = (char**)malloc(sizeof(char*) * (*count));
        for (int i = 0; i < *count; i++) {
            NSURL *url = files[i];
            const char *path = [[url path] UTF8String];
            paths[i] = strdup(path);
        }

        return paths;
    }
}

// Get clipboard text content if any
char* getClipboardText() {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        NSString *text = [pasteboard stringForType:NSPasteboardTypeString];

        if (text == nil) return NULL;

        const char *utf8Text = [text UTF8String];
        return strdup(utf8Text);
    }
}

// Free the file paths array
void freeFilePaths(char **paths, int count) {
    if (!paths) return;
    for (int i = 0; i < count; i++) {
        free(paths[i]);
    }
    free(paths);
}

// Free a single string
void freeString(char *str) {
    if (str) free(str);
}
*/
import "C"
import (
	"unsafe"
)

// CopyFile copies a single file reference to clipboard
func CopyFile(path string) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	C.copyFile(cPath)
}

// CopyFiles copies multiple file references to clipboard
func CopyFiles(paths []string) {
	cPaths := make([]*C.char, len(paths))
	for i, path := range paths {
		cPaths[i] = C.CString(path)
		defer C.free(unsafe.Pointer(cPaths[i]))
	}
	C.copyFiles(&cPaths[0], C.int(len(cPaths)))
}

// CopyText copies text content to clipboard
func CopyText(text string) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	C.copyText(cText)
}

// GetFiles returns file paths currently on clipboard
func GetFiles() []string {
	var count C.int
	cPaths := C.getClipboardFiles(&count)
	if cPaths == nil {
		return nil
	}
	defer C.freeFilePaths(cPaths, count)

	// Convert C array to Go slice
	length := int(count)
	cFiles := (*[1 << 30]*C.char)(unsafe.Pointer(cPaths))[:length:length]

	files := make([]string, length)
	for i := 0; i < length; i++ {
		files[i] = C.GoString(cFiles[i])
	}

	return files
}

// GetText returns text content from clipboard
func GetText() (string, bool) {
	cText := C.getClipboardText()
	if cText == nil {
		return "", false
	}
	defer C.freeString(cText)
	return C.GoString(cText), true
}
