package middleware

import (
	"Rhine-Cloud-Driver/pkg/log"
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	dunno                        = []byte("???")
	centerDot                    = []byte("·")
	dot                          = []byte(".")
	slash                        = []byte("/")
	DefaultErrorWriter io.Writer = os.Stderr
)

const reset string = "\033[0m"

// RecoveryFunc defines the function passable to CustomRecovery.
type RecoveryFunc func(c *gin.Context, err any)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery() gin.HandlerFunc {
	return RecoveryWithWriter(DefaultErrorWriter)
}

// CustomRecovery returns a middleware that recovers from any panics and calls the provided handle func to handle it.
func CustomRecovery(handle RecoveryFunc) gin.HandlerFunc {
	return RecoveryWithWriter(DefaultErrorWriter, handle)
}

// RecoveryWithWriter returns a middleware for a given writer that recovers from any panics and writes a 500 if there was one.
func RecoveryWithWriter(out io.Writer, recovery ...RecoveryFunc) gin.HandlerFunc {
	if len(recovery) > 0 {
		return CustomRecoveryWithWriter(out, recovery[0])
	}
	return CustomRecoveryWithWriter(out, defaultHandleRecovery)
}

func defaultHandleRecovery(c *gin.Context, err any) {
	c.AbortWithStatus(http.StatusInternalServerError)
}

func CustomRecoveryWithWriter(out io.Writer, handle RecoveryFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var buf [4096]byte
				n := runtime.Stack(buf[:], false)
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				stack := stack(3)
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				headers := strings.Split(string(httpRequest), "\r\n")
				for idx, header := range headers {
					current := strings.Split(header, ":")
					if current[0] == "Authorization" {
						headers[idx] = current[0] + ": *"
					}
				}
				//headersToStr := strings.Join(headers, "\r\n")
				if brokenPipe {
					//logger.Printf("%s\n%s%s", err, headersToStr, reset)
					log.Logger.Panic("panic from gin router", zap.Error(err.(error)), zap.Any("header", headers), zap.Any("stack", string(buf[:n])))
				} else if gin.IsDebugging() {
					//logger.Printf("[Recovery] %s panic recovered:\n%s\n%s\n%s%s",
					//	timeFormat(time.Now()), headersToStr, err, stack, reset)
					log.Logger.Panic("panic from gin router[Debugging]", zap.Error(err.(error)), zap.Any("header", headers), zap.Any("stack", string(buf[:n])))
				} else {
					//logger.Printf("[Recovery] %s panic recovered:\n%s\n%s%s",
					//	timeFormat(time.Now()), err, stack, reset)
					log.Logger.Panic("panic from gin router", zap.Error(err.(error)), zap.Any("stack", stack))
				}
				if brokenPipe {
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
				} else {
					handle(c, err)
				}
			}
		}()
		c.Next()
	}
}

// stack returns a nicely formatted stack frame, skipping skip frames.
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contain dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}

// timeFormat returns a customized time string for logger.
func timeFormat(t time.Time) string {
	return t.Format("2006/01/02 - 15:04:05")
}
