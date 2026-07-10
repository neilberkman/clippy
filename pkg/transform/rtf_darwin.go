//go:build darwin && cgo

package transform

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit -framework Foundation

#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>
#include <stdlib.h>
#include <string.h>

static void *clippyMarkdownHTMLToRTF(
    const void *htmlBytes,
    size_t htmlLength,
    size_t *rtfLength,
    char **errorMessage
) {
    @autoreleasepool {
        *rtfLength = 0;
        *errorMessage = NULL;

        NSData *htmlData = [NSData dataWithBytes:htmlBytes length:htmlLength];
        NSDictionary *htmlOptions = @{
            NSDocumentTypeDocumentOption: NSHTMLTextDocumentType,
            NSCharacterEncodingDocumentOption: @(NSUTF8StringEncoding)
        };

        NSError *error = nil;
        NSAttributedString *attributed = [[NSAttributedString alloc]
            initWithData:htmlData
            options:htmlOptions
            documentAttributes:NULL
            error:&error];
        if (attributed == nil) {
            const char *message = error.localizedDescription.UTF8String;
            *errorMessage = strdup(message != NULL ? message : "AppKit could not import HTML");
            return NULL;
        }

        NSDictionary *rtfOptions = @{
            NSDocumentTypeDocumentOption: NSRTFTextDocumentType
        };
        NSData *rtfData = [attributed
            dataFromRange:NSMakeRange(0, attributed.length)
            documentAttributes:rtfOptions
            error:&error];
        if (rtfData == nil) {
            const char *message = error.localizedDescription.UTF8String;
            *errorMessage = strdup(message != NULL ? message : "AppKit could not serialize RTF");
            return NULL;
        }

        void *result = malloc(rtfData.length);
        if (result == NULL && rtfData.length != 0) {
            *errorMessage = strdup("could not allocate RTF buffer");
            return NULL;
        }
        memcpy(result, rtfData.bytes, rtfData.length);
        *rtfLength = rtfData.length;
        return result;
    }
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func materializeRTF(fragmentHTML, _ string) ([]byte, error) {
	documentHTML := `<!doctype html><html><head><meta charset="utf-8"></head><body style="margin: 0;">` + fragmentHTML + `</body></html>`
	htmlBytes := []byte(documentHTML)
	cHTML := C.CBytes(htmlBytes)
	defer C.free(cHTML)

	var length C.size_t
	var errorMessage *C.char
	result := C.clippyMarkdownHTMLToRTF(cHTML, C.size_t(len(htmlBytes)), &length, &errorMessage)
	if errorMessage != nil {
		defer C.free(unsafe.Pointer(errorMessage))
		return nil, fmt.Errorf("%s", C.GoString(errorMessage))
	}
	if result == nil && length != 0 {
		return nil, fmt.Errorf("AppKit returned an empty RTF buffer")
	}
	defer C.free(result)

	return C.GoBytes(result, C.int(length)), nil
}
