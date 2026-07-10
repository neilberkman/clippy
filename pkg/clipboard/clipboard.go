package clipboard

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit -framework CoreServices -framework UniformTypeIdentifiers
#import <Foundation/Foundation.h>
#import <AppKit/NSPasteboard.h>
#import <AppKit/NSPasteboardItem.h>
#import <AppKit/NSApplication.h>
#import <AppKit/NSAttributedString.h>
#import <CoreServices/CoreServices.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>

// Helper function to wait for pasteboard to complete write operation
static int waitForPasteboardChange(NSPasteboard *pasteboard, NSInteger initialChangeCount) {
    NSDate *timeoutDate = [NSDate dateWithTimeIntervalSinceNow:2.0]; // 2-second timeout
    while ([pasteboard changeCount] == initialChangeCount && [timeoutDate timeIntervalSinceNow] > 0) {
        // Run the loop for a very short interval to allow the main thread to process events
        [[NSRunLoop currentRunLoop] runUntilDate:[NSDate dateWithTimeIntervalSinceNow:0.01]];
    }

    // Check if the changeCount was updated
    if ([pasteboard changeCount] == initialChangeCount) {
        return -1; // Timed out
    }

    return 0; // Success
}

// Function to copy a file reference to the clipboard.
//
// Writes a single NSPasteboardItem carrying both public.file-url (so
// receivers attach the file) and public.utf8-plain-text containing the
// basename. The plain-text basename is what makes Messages.app preserve
// the original filename on paste instead of synthesising "FILE_NNNN.<ext>".
// Verified empirically against Messages on macOS 26 (Tahoe); matches the
// shape Finder writes for right-click → Copy.
int copyFile(const char *path) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSString *nsPath = [NSString stringWithUTF8String:path];
        NSURL *fileURL = [NSURL fileURLWithPath:nsPath];
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];

        NSPasteboardItem *item = [[NSPasteboardItem alloc] init];
        [item setString:[fileURL absoluteString] forType:NSPasteboardTypeFileURL];
        [item setString:[nsPath lastPathComponent] forType:NSPasteboardTypeString];

        // Get the current changeCount before operation
        NSInteger initialChangeCount = [pasteboard changeCount];

        // Perform the write operation
        [pasteboard clearContents];
        BOOL success = [pasteboard writeObjects:@[item]];

        if (!success) {
            return -1; // Write operation failed to start
        }

        // Wait for pasteboard to complete
        if (waitForPasteboardChange(pasteboard, initialChangeCount) != 0) {
            return -2; // Timed out
        }

        return 0; // Success
    }
}

// Function to copy multiple file references to the clipboard.
//
// One NSPasteboardItem per file, each with public.file-url and a
// public.utf8-plain-text basename. See copyFile for why the basename is
// included.
int copyFiles(const char **paths, int count) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSMutableArray *items = [NSMutableArray arrayWithCapacity:count];

        for (int i = 0; i < count; i++) {
            NSString *nsPath = [NSString stringWithUTF8String:paths[i]];
            NSURL *fileURL = [NSURL fileURLWithPath:nsPath];
            NSPasteboardItem *item = [[NSPasteboardItem alloc] init];
            [item setString:[fileURL absoluteString] forType:NSPasteboardTypeFileURL];
            [item setString:[nsPath lastPathComponent] forType:NSPasteboardTypeString];
            [items addObject:item];
        }

        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];

        // Get the current changeCount before operation
        NSInteger initialChangeCount = [pasteboard changeCount];

        // Perform the write operation
        [pasteboard clearContents];
        BOOL success = [pasteboard writeObjects:items];

        if (!success) {
            return -1; // Write operation failed to start
        }

        // Wait for pasteboard to complete
        if (waitForPasteboardChange(pasteboard, initialChangeCount) != 0) {
            return -2; // Timed out
        }

        return 0; // Success
    }
}

// Function to copy plain text content to the clipboard
int copyText(const char *text) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSString *nsText = [NSString stringWithUTF8String:text];
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];

        // Get the current changeCount before operation
        NSInteger initialChangeCount = [pasteboard changeCount];

        // Perform the write operation
        [pasteboard clearContents];
        BOOL success = [pasteboard setString:nsText forType:NSPasteboardTypeString];

        if (!success) {
            return -1; // Write operation failed to start
        }

        // Wait for pasteboard to complete
        if (waitForPasteboardChange(pasteboard, initialChangeCount) != 0) {
            return -2; // Timed out
        }

        return 0; // Success
    }
}

// Copy rich text as one pasteboard item with ordered RTF, HTML, and plain-text
// representations. NSData keeps the supplied RTF and UTF-8 bytes intact.
int copyRichText(const void *htmlBytes, size_t htmlLength,
                 const void *rtfBytes, size_t rtfLength,
                 const void *plainTextBytes, size_t plainTextLength) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context

        NSData *rtfData = rtfLength == 0
            ? [NSData data]
            : [NSData dataWithBytes:rtfBytes length:rtfLength];
        NSData *htmlData = htmlLength == 0
            ? [NSData data]
            : [NSData dataWithBytes:htmlBytes length:htmlLength];
        NSData *plainTextData = plainTextLength == 0
            ? [NSData data]
            : [NSData dataWithBytes:plainTextBytes length:plainTextLength];

        NSPasteboardItem *item = [[NSPasteboardItem alloc] init];
        BOOL success = [item setData:rtfData forType:NSPasteboardTypeRTF];
        success = success && [item setData:htmlData forType:NSPasteboardTypeHTML];
        success = success && [item setData:plainTextData forType:NSPasteboardTypeString];

        if (!success) {
            return -1; // Failed to prepare a representation
        }

        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        NSInteger initialChangeCount = [pasteboard changeCount];

        [pasteboard clearContents];
        success = [pasteboard writeObjects:@[item]];

        if (!success) {
            return -1; // Write operation failed to start
        }

        if (waitForPasteboardChange(pasteboard, initialChangeCount) != 0) {
            return -2; // Timed out
        }

        return 0; // Success
    }
}

// Function to copy text with a specific UTI/type to the clipboard
int copyTextWithType(const char *text, const char *typeIdentifier) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSString *nsText = [NSString stringWithUTF8String:text];
        NSString *nsType = [NSString stringWithUTF8String:typeIdentifier];
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];

        // Get the current changeCount before operation
        NSInteger initialChangeCount = [pasteboard changeCount];

        // Perform the write operation
        [pasteboard clearContents];

        // Set both the specific type and plain text for compatibility
        // Many apps expect plain text as a fallback
        BOOL success = [pasteboard setString:nsText forType:nsType];
        if (success) {
            // Also add plain text representation
            [pasteboard setString:nsText forType:NSPasteboardTypeString];
        }

        if (!success) {
            return -1; // Write operation failed to start
        }

        // Wait for pasteboard to complete
        if (waitForPasteboardChange(pasteboard, initialChangeCount) != 0) {
            return -2; // Timed out
        }

        return 0; // Success
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

// Clear the clipboard
int clearClipboard() {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];

        // Get the current changeCount before operation
        NSInteger initialChangeCount = [pasteboard changeCount];

        // Clear the clipboard
        [pasteboard clearContents];

        // Wait for pasteboard to complete
        if (waitForPasteboardChange(pasteboard, initialChangeCount) != 0) {
            return -2; // Timed out
        }

        return 0; // Success
    }
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

// Get preferred file extension for a UTI from macOS's type database
char* getPreferredExtensionForUTI(const char* uti) {
    @autoreleasepool {
        NSString *utiString = [NSString stringWithUTF8String:uti];

        if (@available(macOS 11.0, *)) {
            UTType *utType = [UTType typeWithIdentifier:utiString];
            NSString *ext = utType.preferredFilenameExtension;

            if (ext) {
                return strdup([ext UTF8String]);
            }
        } else {
            // Fallback to deprecated API
            #pragma clang diagnostic push
            #pragma clang diagnostic ignored "-Wdeprecated-declarations"
            CFStringRef utiRef = CFStringCreateWithCString(NULL, uti, kCFStringEncodingUTF8);
            CFStringRef extRef = UTTypeCopyPreferredTagWithClass(utiRef, kUTTagClassFilenameExtension);

            if (extRef) {
                const char *ext = CFStringGetCStringPtr(extRef, kCFStringEncodingUTF8);
                char *result = ext ? strdup(ext) : NULL;
                CFRelease(extRef);
                CFRelease(utiRef);
                return result;
            }
            CFRelease(utiRef);
            #pragma clang diagnostic pop
        }

        return NULL;
    }
}

// Save RTFD data from clipboard to directory bundle
// Returns 0 on success, -1 on error
int saveRTFDToPath(const char* data, int length, const char* path) {
    @autoreleasepool {
        // Deserialize the flat-rtfd data
        NSData *rtfdData = [NSData dataWithBytes:data length:length];

        NSError *error = nil;
        NSAttributedString *attrString = [[NSAttributedString alloc]
            initWithData:rtfdData
            options:@{NSDocumentTypeDocumentOption: NSRTFDTextDocumentType}
            documentAttributes:nil
            error:&error];

        if (!attrString || error) {
            if (error) {
                NSLog(@"Failed to create attributed string: %@", error);
            }
            return -1;
        }

        // Create the .rtfd bundle directory
        NSString *rtfdPath = [NSString stringWithUTF8String:path];
        NSFileManager *fm = [NSFileManager defaultManager];

        // Remove existing if present
        [fm removeItemAtPath:rtfdPath error:nil];

        // Create the RTFD wrapper directly from the attributed string
        NSFileWrapper *wrapper = [attrString fileWrapperFromRange:NSMakeRange(0, [attrString length])
                                              documentAttributes:@{NSDocumentTypeDocumentOption: NSRTFDTextDocumentType}
                                              error:&error];
        if (!wrapper || error) {
            if (error) {
                NSLog(@"Failed to create file wrapper: %@", error);
            }
            return -1;
        }

        // Write the wrapper to disk
        BOOL written = [wrapper writeToURL:[NSURL fileURLWithPath:rtfdPath]
                                   options:0
                                   originalContentsURL:nil
                                   error:&error];

        if (!written || error) {
            if (error) {
                NSLog(@"Failed to write file wrapper: %@", error);
            }
            return -1;
        }

        return 0;
    }
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// CopyFile copies a single file reference to clipboard
func CopyFile(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	result := C.copyFile(cPath)

	switch result {
	case 0:
		return nil
	case -1:
		return fmt.Errorf("failed to write to clipboard")
	case -2:
		return fmt.Errorf("clipboard operation timed out")
	default:
		return fmt.Errorf("unknown clipboard error: %d", result)
	}
}

// CopyFiles copies multiple file references to clipboard
func CopyFiles(paths []string) error {
	cPaths := make([]*C.char, len(paths))
	for i, path := range paths {
		cPaths[i] = C.CString(path)
		defer C.free(unsafe.Pointer(cPaths[i]))
	}
	result := C.copyFiles(&cPaths[0], C.int(len(cPaths)))

	switch result {
	case 0:
		return nil
	case -1:
		return fmt.Errorf("failed to write to clipboard")
	case -2:
		return fmt.Errorf("clipboard operation timed out")
	default:
		return fmt.Errorf("unknown clipboard error: %d", result)
	}
}

// CopyText copies text content to clipboard
func CopyText(text string) error {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	result := C.copyText(cText)

	switch result {
	case 0:
		return nil
	case -1:
		return fmt.Errorf("failed to write to clipboard")
	case -2:
		return fmt.Errorf("clipboard operation timed out")
	default:
		return fmt.Errorf("unknown clipboard error: %d", result)
	}
}

// CopyRichText copies ordered RTF, HTML, and plain-text representations to a
// single pasteboard item. HTML and plainText are encoded as UTF-8.
func CopyRichText(htmlContent string, rtfContent []byte, plainText string) error {
	htmlBytes := []byte(htmlContent)
	plainTextBytes := []byte(plainText)

	cHTML := cBytes(htmlBytes)
	if cHTML != nil {
		defer C.free(cHTML)
	}
	cRTF := cBytes(rtfContent)
	if cRTF != nil {
		defer C.free(cRTF)
	}
	cPlainText := cBytes(plainTextBytes)
	if cPlainText != nil {
		defer C.free(cPlainText)
	}

	result := C.copyRichText(
		cHTML,
		C.size_t(len(htmlBytes)),
		cRTF,
		C.size_t(len(rtfContent)),
		cPlainText,
		C.size_t(len(plainTextBytes)),
	)

	switch result {
	case 0:
		return nil
	case -1:
		return fmt.Errorf("failed to write to clipboard")
	case -2:
		return fmt.Errorf("clipboard operation timed out")
	default:
		return fmt.Errorf("unknown clipboard error: %d", result)
	}
}

func cBytes(data []byte) unsafe.Pointer {
	if len(data) == 0 {
		return nil
	}
	return C.CBytes(data)
}

// CopyTextWithType copies text with a specific UTI type to clipboard
// Common types: "public.html", "public.json", "public.xml", "public.plain-text"
func CopyTextWithType(text string, typeIdentifier string) error {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	cType := C.CString(typeIdentifier)
	defer C.free(unsafe.Pointer(cType))
	result := C.copyTextWithType(cText, cType)

	switch result {
	case 0:
		return nil
	case -1:
		return fmt.Errorf("failed to write to clipboard")
	case -2:
		return fmt.Errorf("clipboard operation timed out")
	default:
		return fmt.Errorf("unknown clipboard error: %d", result)
	}
}

// Clear clears the clipboard
func Clear() error {
	result := C.clearClipboard()

	switch result {
	case 0:
		return nil
	case -2:
		return fmt.Errorf("clipboard operation timed out")
	default:
		return fmt.Errorf("unknown clipboard error: %d", result)
	}
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

// GetPreferredExtensionForUTI returns the preferred file extension for a UTI
// using macOS's canonical type database. Returns empty string if not found.
// Example: "public.png" -> "png", "public.jpeg" -> "jpg"
func GetPreferredExtensionForUTI(uti string) string {
	cUTI := C.CString(uti)
	defer C.free(unsafe.Pointer(cUTI))

	cExt := C.getPreferredExtensionForUTI(cUTI)
	if cExt == nil {
		return ""
	}
	defer C.freeString(cExt)

	return C.GoString(cExt)
}

// SaveRTFDToPath saves RTFD clipboard data to an .rtfd bundle on disk
// The path should end with .rtfd extension
// Returns an error if the save fails
func SaveRTFDToPath(data []byte, path string) error {
	if len(data) == 0 {
		return fmt.Errorf("empty RTFD data")
	}

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	cData := (*C.char)(unsafe.Pointer(&data[0]))
	length := C.int(len(data))

	result := C.saveRTFDToPath(cData, length, cPath)
	if result != 0 {
		return fmt.Errorf("failed to save RTFD bundle")
	}

	return nil
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

	// Priority 2: Check for rich UTI types on clipboard (images, etc.)
	// This must come BEFORE text check because browsers put both image data
	// and URL text on clipboard - we want the image data
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

	// Priority 3: Check for text content (fallback)
	// This comes last so image data takes precedence over accompanying URLs
	if text, ok := GetText(); ok {
		return &ClipboardContent{
			Type:   "public.utf8-plain-text",
			Data:   []byte(text),
			IsText: true,
		}, nil
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
		"com.apple.flat-rtfd", // RTF with embedded images/attachments (priority)
		"public.rtf",          // Plain RTF formatting
		"com.apple.rtfd",      // RTFD bundle
		"public.pdf",
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
