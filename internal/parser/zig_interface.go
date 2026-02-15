package parser

/*
#cgo linux,amd64 LDFLAGS: ${SRCDIR}/native/libparser_linux_amd64.a
#cgo darwin,arm64 LDFLAGS: ${SRCDIR}/native/libparser_darwin_arm64.a
#include "native/parser.h"
*/
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"github.com/alextreichler/bundleViewer/internal/logutil"
	"github.com/alextreichler/bundleViewer/internal/models"
)

func init() {
	logutil.FingerprinterOverride = ZigFingerprint
}

type ZigParser struct{}

func (zp *ZigParser) Parse(data []byte) ([]C.LogLineInfo, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var count C.size_t
	resPtr := C.zig_parse_logs((*C.char)(unsafe.Pointer(&data[0])), C.size_t(len(data)), &count)
	if resPtr == nil {
		return nil, fmt.Errorf("zig_parse_logs failed")
	}

	// Create a slice that points to the C-allocated memory
	lines := unsafe.Slice(resPtr, int(count))

	return lines, nil
}

func (zp *ZigParser) Free(lines []C.LogLineInfo) {
	if len(lines) == 0 {
		return
	}
	C.zig_free_result((*C.LogLineInfo)(unsafe.Pointer(&lines[0])), C.size_t(len(lines)))
}

var fingerprintBufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

func ZigFingerprint(message string) string {
	if len(message) == 0 {
		return ""
	}

	msgLen := len(message)
	// Markers like <NUM> can be longer than the tokens they replace.
	// 2x should be plenty.
	needed := msgLen * 2
	
	var buf []byte
	if needed <= 4096 {
		buf = fingerprintBufPool.Get().([]byte)
		defer fingerprintBufPool.Put(buf)
	} else {
		buf = make([]byte, needed)
	}

	msgPtr := unsafe.StringData(message)
	n := C.zig_fingerprint((*C.char)(unsafe.Pointer(msgPtr)), C.size_t(msgLen), (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(len(buf)))

	return string(buf[:n])
}

func (lp *LogParser) parseFileZig(filePath string, batchChan chan<- []*models.LogEntry) error {
	if lp.tracker != nil {
		lp.tracker.Update(1, fmt.Sprintf("Parsing logs (Zig): %s", filepath.Base(filePath)))
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	zp := &ZigParser{}
	lines, err := zp.Parse(data)
	if err != nil {
		return err
	}
	defer zp.Free(lines)

	nodeName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	if filepath.Base(filePath) == "redpanda.log" {
		nodeName = "redpanda"
	}

	batch := lp.getBatch()
	var currentEntry *models.LogEntry

	for i, lineInfo := range lines {
		if lp.tracker != nil && i%100000 == 0 {
			pct := (i * 100) / len(lines)
			lp.tracker.SetStatus(fmt.Sprintf("Parsing logs (Zig): %s [%d/%d lines, %d%%]", filepath.Base(filePath), i, len(lines), pct))
		}

		lineStart := int(lineInfo.line_start)
		lineLen := int(lineInfo.line_len)
		if lineStart+lineLen > len(data) {
			continue
		}
		lineBytes := data[lineStart : lineStart+lineLen]
		line := string(lineBytes)

		if lineInfo.level_len > 0 {
			// Fast path: Zig already found level and timestamp
			// We still need shard, component and message which Zig didn't extract fully
			// But we can use the existing Go logic with the info we have
			
			// For simplicity, let's just use tryParseStandardLog for now, 
			// but in a real "fast" version we would use the offsets directly.
			// However, Zig already validated it looks like a standard log.
			
			entry, ok := lp.parseStandardLine(line, nodeName, filePath, i+1)
			if ok {
				if currentEntry != nil {
					batch = append(batch, currentEntry)
					if len(batch) >= lp.batchSize {
						batchChan <- batch
						batch = lp.getBatch()
					}
				}
				currentEntry = entry
				continue
			}
		}

		// Fallback for non-standard lines or if Zig didn't find info
		var entry *models.LogEntry
		var ok bool

		// Check if it's K8s/JSON
		entry, ok = lp.parseK8sLine(line, nodeName, filePath, i+1)
		if !ok {
			entry, ok = lp.parseJSONLine(line, nodeName, filePath, i+1)
		}

		if ok {
			if currentEntry != nil {
				batch = append(batch, currentEntry)
				if len(batch) >= lp.batchSize {
					batchChan <- batch
					batch = lp.getBatch()
				}
			}
			currentEntry = entry
		} else if currentEntry != nil {
			// Multiline append
			currentEntry.Message += "\n" + line
			currentEntry.Raw += "\n" + line
		} else {
			// Fallback for first line
			currentEntry = lp.getEntry()
			currentEntry.Level = "INFO"
			currentEntry.Node = nodeName
			currentEntry.Component = "stdout"
			currentEntry.Message = line
			currentEntry.Raw = line
			currentEntry.LineNumber = i + 1
			currentEntry.FilePath = filePath
		}
	}

	if currentEntry != nil {
		batch = append(batch, currentEntry)
	}
	if len(batch) > 0 {
		batchChan <- batch
	}

	return nil
}
