//go:build darwin

package spotlight

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreServices -framework Foundation
#import <CoreServices/CoreServices.h>
#import <Foundation/Foundation.h>

// FileItem represents a file with its modification date from Spotlight
typedef struct {
	char* path;
	double modTime; // CFAbsoluteTime
} FileItem;

// searchFiles performs a Spotlight search and returns matching file paths with mod times
FileItem* searchFiles(const char* query, int* resultCount, int maxResults) {
	@autoreleasepool {
		NSString *queryStr = [NSString stringWithUTF8String:query];

		// Build base filename query
		NSString *nameQuery;
		if ([queryStr hasPrefix:@"."]) {
			// Extension search: ".pdf" -> files ending with .pdf
			nameQuery = [NSString stringWithFormat:@"kMDItemFSName == '*%@'cd", queryStr];
		} else {
			// Substring search: "invoice" or "report.xlsx" -> files containing the string
			nameQuery = [NSString stringWithFormat:@"kMDItemFSName == '*%@*'cd", queryStr];
		}

		// Add date filter: only files modified in last 90 days
		// This dramatically reduces the result set at the Spotlight level
		NSString *queryFormat = [NSString stringWithFormat:@"%@ && kMDItemContentModificationDate >= $time.today(-90)", nameQuery];

		MDQueryRef mdQuery = MDQueryCreate(kCFAllocatorDefault, (__bridge CFStringRef)queryFormat, NULL, NULL);

		if (!mdQuery) {
			*resultCount = 0;
			return NULL;
		}

		// Note: We sort results in Go after fetching
		// MDQuery sorting APIs are unreliable

		// Execute the query synchronously
		Boolean success = MDQueryExecute(mdQuery, kMDQuerySynchronous);
		if (!success) {
			CFRelease(mdQuery);
			*resultCount = 0;
			return NULL;
		}

		// Get result count
		CFIndex count = MDQueryGetResultCount(mdQuery);
		if (count == 0) {
			CFRelease(mdQuery);
			*resultCount = 0;
			return NULL;
		}

		// Limit results
		if (maxResults > 0 && count > maxResults) {
			count = maxResults;
		}

		// Allocate array for results
		FileItem *results = (FileItem *)malloc(sizeof(FileItem) * count);
		int actualCount = 0;

		// Get file paths and modification times from Spotlight
		for (CFIndex i = 0; i < count; i++) {
			MDItemRef item = (MDItemRef)MDQueryGetResultAtIndex(mdQuery, i);
			if (!item) continue;

			// Get path
			CFStringRef pathRef = MDItemCopyAttribute(item, kMDItemPath);
			if (!pathRef) continue;

			const char *pathCStr = CFStringGetCStringPtr(pathRef, kCFStringEncodingUTF8);
			char buffer[4096];
			if (!pathCStr) {
				// If direct pointer fails, use buffer
				if (CFStringGetCString(pathRef, buffer, sizeof(buffer), kCFStringEncodingUTF8)) {
					pathCStr = buffer;
				}
			}

			// Get modification date from Spotlight
			CFDateRef modDateRef = MDItemCopyAttribute(item, kMDItemContentModificationDate);
			double modTime = 0.0;
			if (modDateRef) {
				modTime = CFDateGetAbsoluteTime(modDateRef);
				CFRelease(modDateRef);
			}

			if (pathCStr) {
				results[actualCount].path = strdup(pathCStr);
				results[actualCount].modTime = modTime;
				actualCount++;
			}

			CFRelease(pathRef);
		}

		CFRelease(mdQuery);
		*resultCount = actualCount;
		return results;
	}
}

// freeResults frees the memory allocated by searchFiles
void freeResults(FileItem* results, int count) {
	for (int i = 0; i < count; i++) {
		free(results[i].path);
	}
	free(results);
}
*/
import "C"
import (
	"fmt"
	"os"
	"sort"
	"time"
	"unsafe"
)

// SearchOptions configures Spotlight search behavior
type SearchOptions struct {
	Query      string   // Search query (filename pattern)
	Scope      []string // Optional: limit to specific directories (not implemented yet)
	MaxResults int      // Optional: limit result count (0 = no limit)
}

// FileResult represents a file found by Spotlight
type FileResult struct {
	Path string // Full path to the file
	Name string // Filename only (extracted from path)
}

// FileInfo represents a file with full metadata (compatible with recent.FileInfo)
type FileInfo struct {
	Path     string
	Name     string
	Size     int64
	Modified time.Time
	IsDir    bool
}

// cfAbsoluteTimeToGoTime converts CFAbsoluteTime to Go time.Time
// CFAbsoluteTime is seconds since 2001-01-01 00:00:00 UTC
// Unix epoch is 1970-01-01 00:00:00 UTC
// Difference is 978307200 seconds
func cfAbsoluteTimeToGoTime(cfTime float64) time.Time {
	const cfAbsoluteTimeToUnixEpoch = 978307200
	unixTime := int64(cfTime) + cfAbsoluteTimeToUnixEpoch
	return time.Unix(unixTime, 0)
}

// Search performs a Spotlight search for files matching the query
func Search(opts SearchOptions) ([]FileResult, error) {
	if opts.Query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	maxResults := opts.MaxResults
	if maxResults == 0 {
		maxResults = 100 // Default limit to prevent overwhelming results
	}

	cQuery := C.CString(opts.Query)
	defer C.free(unsafe.Pointer(cQuery))

	var resultCount C.int
	cResults := C.searchFiles(cQuery, &resultCount, C.int(maxResults))

	if cResults == nil || resultCount == 0 {
		return []FileResult{}, nil // No results found
	}
	defer C.freeResults(cResults, resultCount)

	// Convert C array to Go slice
	results := make([]FileResult, int(resultCount))
	cResultsSlice := (*[1 << 28]C.FileItem)(unsafe.Pointer(cResults))[:resultCount:resultCount]

	for i := 0; i < int(resultCount); i++ {
		path := C.GoString(cResultsSlice[i].path)
		results[i] = FileResult{
			Path: path,
			Name: extractFilename(path),
		}
	}

	return results, nil
}

// SearchWithMetadata performs a Spotlight search and enriches results with file metadata
// This is the high-level business function that returns files ready for use
// Results are sorted by modification time (most recent first)
func SearchWithMetadata(opts SearchOptions) ([]FileInfo, error) {
	if opts.Query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	maxResults := opts.MaxResults
	if maxResults == 0 {
		maxResults = 100 // Default limit to prevent overwhelming results
	}

	cQuery := C.CString(opts.Query)
	defer C.free(unsafe.Pointer(cQuery))

	var resultCount C.int
	cResults := C.searchFiles(cQuery, &resultCount, C.int(maxResults))

	if cResults == nil || resultCount == 0 {
		return []FileInfo{}, nil // No results found
	}
	defer C.freeResults(cResults, resultCount)

	// Convert C array to Go slice with full metadata
	cResultsSlice := (*[1 << 28]C.FileItem)(unsafe.Pointer(cResults))[:resultCount:resultCount]
	var files []FileInfo

	for i := 0; i < int(resultCount); i++ {
		path := C.GoString(cResultsSlice[i].path)
		modTime := cfAbsoluteTimeToGoTime(float64(cResultsSlice[i].modTime))

		// Get size and IsDir from os.Stat (these aren't available from Spotlight)
		info, err := os.Stat(path)
		if err != nil {
			// Skip files that can't be accessed
			continue
		}

		files = append(files, FileInfo{
			Path:     path,
			Name:     extractFilename(path),
			Size:     info.Size(),
			Modified: modTime, // Use modification time from Spotlight
			IsDir:    info.IsDir(),
		})
	}

	// Sort by modification time, most recent first
	sort.Slice(files, func(i, j int) bool {
		return files[i].Modified.After(files[j].Modified)
	})

	return files, nil
}

// extractFilename extracts the filename from a full path
func extractFilename(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
