//go:build darwin

package transform

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit
#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>

// convertMarkdownToRTF converts markdown text to RTF data
// Returns pointer to RTF bytes and length, or NULL/0 on error
typedef struct {
	void* bytes;
	size_t length;
} RTFResult;

RTFResult convertMarkdownToRTF(const char* markdown) {
	RTFResult result = {NULL, 0};

	@autoreleasepool {
		NSString *markdownStr = [NSString stringWithUTF8String:markdown];
		if (!markdownStr) {
			return result;
		}

		NSError *error = nil;
		NSAttributedString *attrStr = [[NSAttributedString alloc]
			initWithMarkdown:[markdownStr dataUsingEncoding:NSUTF8StringEncoding]
			options:nil
			baseURL:nil
			error:&error];

		if (error || !attrStr) {
			return result;
		}

		// Convert NSAttributedString to RTF data
		NSRange range = NSMakeRange(0, [attrStr length]);
		NSDictionary *attributes = @{NSDocumentTypeDocumentAttribute: NSRTFTextDocumentType};
		NSData *rtfData = [attrStr dataFromRange:range
			documentAttributes:attributes
			error:&error];

		if (error || !rtfData) {
			return result;
		}

		// Copy the bytes so they survive autorelease pool
		result.length = [rtfData length];
		result.bytes = malloc(result.length);
		if (result.bytes) {
			memcpy(result.bytes, [rtfData bytes], result.length);
		} else {
			result.length = 0;
		}
	}

	return result;
}

// freeRTFResult releases RTF result memory
void freeRTFResult(RTFResult result) {
	if (result.bytes) {
		free(result.bytes);
	}
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// MarkdownToRTF converts markdown text to RTF format
// Returns RTF data as bytes
func MarkdownToRTF(markdown string) ([]byte, error) {
	cMarkdown := C.CString(markdown)
	defer C.free(unsafe.Pointer(cMarkdown))

	result := C.convertMarkdownToRTF(cMarkdown)
	if result.bytes == nil || result.length == 0 {
		return nil, fmt.Errorf("failed to convert markdown to RTF")
	}
	defer C.freeRTFResult(result)

	// Copy RTF data to Go bytes
	rtfBytes := C.GoBytes(result.bytes, C.int(result.length))
	return rtfBytes, nil
}
