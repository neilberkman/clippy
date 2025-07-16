package clipboard

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit -framework CoreServices -framework UniformTypeIdentifiers
#import <Foundation/Foundation.h>
#import <AppKit/NSPasteboard.h>
#import <AppKit/NSApplication.h>
#import <CoreServices/CoreServices.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>

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

// Get UTI for a file path
char* getUTIForFile(const char* path) {
    @autoreleasepool {
        CFStringRef pathRef = CFStringCreateWithCString(NULL, path, kCFStringEncodingUTF8);
        CFURLRef urlRef = CFURLCreateWithFileSystemPath(NULL, pathRef, kCFURLPOSIXPathStyle, false);
        CFRelease(pathRef);
        if (urlRef == NULL) return NULL;

        CFStringRef utiRef = NULL;
        CFStringRef pathExtension = CFURLCopyPathExtension(urlRef);
        if (pathExtension) {
            // Use the newer API when available, fallback to deprecated API
            if (@available(macOS 11.0, *)) {
                NSString *extension = (__bridge NSString*)pathExtension;
                UTType *utType = [UTType typeWithFilenameExtension:extension];
                if (utType) {
                    utiRef = CFStringCreateCopy(NULL, (__bridge CFStringRef)utType.identifier);
                }
            } else {
                #pragma clang diagnostic push
                #pragma clang diagnostic ignored "-Wdeprecated-declarations"
                utiRef = UTTypeCreatePreferredIdentifierForTag(kUTTagClassFilenameExtension,
                                                              pathExtension,
                                                              NULL);
                #pragma clang diagnostic pop
            }
            CFRelease(pathExtension);
        }
        CFRelease(urlRef);

        if (utiRef == NULL) return NULL;

        const char* uti = CFStringGetCStringPtr(utiRef, kCFStringEncodingUTF8);
        char* result = uti ? strdup(uti) : NULL;
        CFRelease(utiRef);
        return result;
    }
}

// Get available types on clipboard
char** getClipboardTypes(int *count) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        NSArray *types = [pasteboard types];

        *count = (int)[types count];
        if (*count == 0) return NULL;

        char **typeStrings = (char**)malloc(sizeof(char*) * (*count));
        for (int i = 0; i < *count; i++) {
            NSString *type = types[i];
            const char *typeStr = [type UTF8String];
            typeStrings[i] = strdup(typeStr);
        }

        return typeStrings;
    }
}

// Get clipboard data for a specific type
char* getClipboardDataForType(const char* type, int *length) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        NSString *typeString = [NSString stringWithUTF8String:type];
        NSData *data = [pasteboard dataForType:typeString];

        if (data == nil) {
            *length = 0;
            return NULL;
        }

        *length = (int)[data length];
        char *result = (char*)malloc(*length);
        [data getBytes:result length:*length];
        return result;
    }
}

// Check if clipboard contains a specific type
int clipboardContainsType(const char* type) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        NSString *typeString = [NSString stringWithUTF8String:type];
        NSArray *types = [pasteboard types];
        return [types containsObject:typeString] ? 1 : 0;
    }
}

// Check if a UTI conforms to a parent type (e.g., check if UTI is text)
int utiConformsTo(const char* uti, const char* parentType) {
    @autoreleasepool {
        CFStringRef utiRef = CFStringCreateWithCString(NULL, uti, kCFStringEncodingUTF8);
        CFStringRef parentRef = CFStringCreateWithCString(NULL, parentType, kCFStringEncodingUTF8);

        int result = 0;

        // Use the newer API when available
        if (@available(macOS 11.0, *)) {
            NSString *utiString = (__bridge NSString*)utiRef;
            NSString *parentString = (__bridge NSString*)parentRef;
            UTType *utType = [UTType typeWithIdentifier:utiString];
            UTType *parentUTType = [UTType typeWithIdentifier:parentString];

            if (utType && parentUTType) {
                result = [utType conformsToType:parentUTType] ? 1 : 0;
            }
        } else {
            // Fallback to deprecated API
            #pragma clang diagnostic push
            #pragma clang diagnostic ignored "-Wdeprecated-declarations"
            result = UTTypeConformsTo(utiRef, parentRef) ? 1 : 0;
            #pragma clang diagnostic pop
        }

        CFRelease(utiRef);
        CFRelease(parentRef);
        return result;
    }
}
*/
import "C"
import (
	"fmt"
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

// GetUTIForFile returns the UTI (Uniform Type Identifier) for a file path
func GetUTIForFile(path string) (string, bool) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	cUTI := C.getUTIForFile(cPath)
	if cUTI == nil {
		return "", false
	}
	defer C.freeString(cUTI)

	return C.GoString(cUTI), true
}

// GetClipboardTypes returns all available types on clipboard
func GetClipboardTypes() []string {
	var count C.int
	cTypes := C.getClipboardTypes(&count)
	if cTypes == nil {
		return nil
	}
	defer C.freeFilePaths(cTypes, count)

	// Convert C array to Go slice
	length := int(count)
	cTypeArray := (*[1 << 30]*C.char)(unsafe.Pointer(cTypes))[:length:length]

	types := make([]string, length)
	for i := 0; i < length; i++ {
		types[i] = C.GoString(cTypeArray[i])
	}

	return types
}

// GetClipboardDataForType returns data for a specific type from clipboard
func GetClipboardDataForType(typeStr string) ([]byte, bool) {
	cType := C.CString(typeStr)
	defer C.free(unsafe.Pointer(cType))

	var length C.int
	cData := C.getClipboardDataForType(cType, &length)
	if cData == nil {
		return nil, false
	}
	defer C.free(unsafe.Pointer(cData))

	// Convert C data to Go byte slice
	data := C.GoBytes(unsafe.Pointer(cData), length)
	return data, true
}

// ContainsType checks if clipboard contains a specific type
func ContainsType(typeStr string) bool {
	cType := C.CString(typeStr)
	defer C.free(unsafe.Pointer(cType))

	return C.clipboardContainsType(cType) == 1
}

// UTIConformsTo checks if a UTI conforms to a parent type using macOS UTI system
func UTIConformsTo(uti, parentType string) bool {
	cUTI := C.CString(uti)
	defer C.free(unsafe.Pointer(cUTI))

	cParent := C.CString(parentType)
	defer C.free(unsafe.Pointer(cParent))

	return C.utiConformsTo(cUTI, cParent) == 1
}

// ClipboardContent represents the content and type information from clipboard
type ClipboardContent struct {
	Type     string // UTI or MIME type
	Data     []byte // Raw data
	IsText   bool   // Whether this is text content
	IsFile   bool   // Whether this is file reference
	FilePath string // File path if IsFile is true
}

// GetClipboardContent returns clipboard content with smart type detection
// Uses hybrid approach: UTI -> MIME -> mimetype fallback
func GetClipboardContent() (*ClipboardContent, error) {
	// Priority 1: Check for file URLs (highest reliability)
	if files := GetFiles(); len(files) > 0 {
		// For multiple files, just return info about the first one
		filePath := files[0]
		uti, _ := GetUTIForFile(filePath)
		return &ClipboardContent{
			Type:     uti,
			IsFile:   true,
			FilePath: filePath,
		}, nil
	}

	// Priority 2: Check for text content
	if text, ok := GetText(); ok {
		return &ClipboardContent{
			Type:   "public.utf8-plain-text",
			Data:   []byte(text),
			IsText: true,
		}, nil
	}

	// Priority 3: Check for rich UTI types on clipboard
	types := GetClipboardTypes()
	for _, typeStr := range types {
		// Look for specific image types first
		if isImageUTI(typeStr) {
			if data, ok := GetClipboardDataForType(typeStr); ok {
				return &ClipboardContent{
					Type:   typeStr,
					Data:   data,
					IsText: false,
				}, nil
			}
		}

		// Look for other rich content types
		if isRichContentUTI(typeStr) {
			if data, ok := GetClipboardDataForType(typeStr); ok {
				return &ClipboardContent{
					Type:   typeStr,
					Data:   data,
					IsText: false,
				}, nil
			}
		}
	}

	// Priority 4: Check for generic types like public.data
	for _, typeStr := range types {
		if typeStr == "public.data" || typeStr == "public.content" {
			if data, ok := GetClipboardDataForType(typeStr); ok {
				// Use mimetype detection as fallback
				return &ClipboardContent{
					Type:   typeStr,
					Data:   data,
					IsText: false,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("no supported content found on clipboard")
}

// isImageUTI checks if a UTI represents an image type
func isImageUTI(uti string) bool {
	imageUTIs := []string{
		"public.png",
		"public.jpeg",
		"public.tiff",
		"public.gif",
		"public.bmp",
		"public.webp",
		"public.heic",
		"public.svg-image",
	}

	for _, imageUTI := range imageUTIs {
		if uti == imageUTI {
			return true
		}
	}

	return false
}

// isRichContentUTI checks if a UTI represents rich content
func isRichContentUTI(uti string) bool {
	richUTIs := []string{
		"public.pdf",
		"public.rtf",
		"public.html",
		"public.xml",
		"public.json",
		"public.zip-archive",
		"public.tar-archive",
		"public.mp3",
		"public.mp4",
		"public.mpeg-4",
		"public.quicktime-movie",
	}

	for _, richUTI := range richUTIs {
		if uti == richUTI {
			return true
		}
	}

	return false
}
