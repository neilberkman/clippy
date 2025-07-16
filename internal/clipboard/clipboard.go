package clipboard

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit
#import <Foundation/Foundation.h>
#import <AppKit/NSPasteboard.h>
#import <AppKit/NSApplication.h>
#import <AppKit/NSBitmapImageRep.h>
#import <CoreServices/CoreServices.h>

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

// Get clipboard data for any type
void* getClipboardData(const char *type, int *length) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        NSString *nsType = [NSString stringWithUTF8String:type];
        NSData *data = [pasteboard dataForType:nsType];

        if (data == nil) {
            *length = 0;
            return NULL;
        }

        *length = (int)[data length];
        void *buffer = malloc(*length);
        [data getBytes:buffer length:*length];
        return buffer;
    }
}

// Check what types are available on clipboard
char** getClipboardTypes(int *count) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        NSArray *types = [pasteboard types];

        *count = (int)[types count];
        if (*count == 0) return NULL;

        char **typeList = (char**)malloc(sizeof(char*) * (*count));
        for (int i = 0; i < *count; i++) {
            NSString *type = types[i];
            const char *typeStr = [type UTF8String];
            typeList[i] = strdup(typeStr);
        }

        return typeList;
    }
}

// Check if clipboard has image data
int hasClipboardImage() {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        NSArray *types = [pasteboard types];

        for (NSString *type in types) {
            if ([type isEqualToString:NSPasteboardTypePNG] ||
                [type isEqualToString:NSPasteboardTypeTIFF] ||
                [type hasPrefix:@"public."] && [type containsString:@"image"]) {
                return 1;
            }
        }
        return 0;
    }
}

// Get image data from clipboard
void* getClipboardImage(int *length) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];

        // Try PNG first
        NSData *data = [pasteboard dataForType:NSPasteboardTypePNG];

        // Try TIFF if no PNG
        if (data == nil) {
            data = [pasteboard dataForType:NSPasteboardTypeTIFF];
            if (data != nil) {
                // Convert TIFF to PNG
                NSBitmapImageRep *imageRep = [NSBitmapImageRep imageRepWithData:data];
                data = [imageRep representationUsingType:NSBitmapImageFileTypePNG properties:@{}];
            }
        }

        if (data == nil) {
            *length = 0;
            return NULL;
        }

        *length = (int)[data length];
        void *buffer = malloc(*length);
        [data getBytes:buffer length:*length];
        return buffer;
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

// Free data buffer
void freeData(void *data) {
    if (data) free(data);
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

// GetTypes returns available types on clipboard
func GetTypes() []string {
	var count C.int
	cTypes := C.getClipboardTypes(&count)
	if cTypes == nil {
		return nil
	}
	defer C.freeFilePaths(cTypes, count) // Reuse same free function

	// Convert C array to Go slice
	length := int(count)
	cTypeArray := (*[1 << 30]*C.char)(unsafe.Pointer(cTypes))[:length:length]

	types := make([]string, length)
	for i := 0; i < length; i++ {
		types[i] = C.GoString(cTypeArray[i])
	}

	return types
}

// HasImage checks if clipboard contains image data
func HasImage() bool {
	return C.hasClipboardImage() != 0
}

// GetImage returns image data from clipboard as PNG bytes
func GetImage() ([]byte, bool) {
	var length C.int
	data := C.getClipboardImage(&length)
	if data == nil {
		return nil, false
	}
	defer C.freeData(data)

	// Copy C data to Go slice
	size := int(length)
	result := make([]byte, size)
	copy(result, (*[1 << 30]byte)(data)[:size:size])

	return result, true
}

// GetData returns raw clipboard data for a specific type
func GetData(dataType string) ([]byte, bool) {
	var length C.int
	cType := C.CString(dataType)
	defer C.free(unsafe.Pointer(cType))

	data := C.getClipboardData(cType, &length)
	if data == nil {
		return nil, false
	}
	defer C.freeData(data)

	// Copy C data to Go slice
	size := int(length)
	result := make([]byte, size)
	copy(result, (*[1 << 30]byte)(data)[:size:size])

	return result, true
}
