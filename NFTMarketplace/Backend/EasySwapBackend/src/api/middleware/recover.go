package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"

	"github.com/ProjectsTask/EasySwapBase/errcode"
	"github.com/ProjectsTask/EasySwapBase/logger/xzap"
	"github.com/ProjectsTask/EasySwapBase/xhttp"
	"github.com/gin-gonic/gin"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// RecoverMiddleware 是一个Gin中间件，用于捕获并处理在处理HTTP请求过程中发生的panic。
// 当发生panic时，它会记录错误信息，包括请求内容和堆栈跟踪，并返回一个通用的错误响应给客户端。
// 返回值是一个 gin.HandlerFunc 类型的函数，用于处理HTTP请求。
func RecoverMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用defer确保在函数返回时执行恢复逻辑
		defer func() {
			// 尝试从panic中恢复
			if cause := recover(); cause != nil {
				// 记录错误信息，包括请求内容、panic原因和堆栈跟踪
				xzap.WithContext(c.Request.Context()).Errorf("[Recovery] panic recovered, request:%s%v [## stack:]:\n%s", dumpRequest(c.Request), cause, dumpStack(3))
				// 返回一个通用的错误响应给客户端
				xhttp.Error(c, errcode.ErrUnexpected)
			}
		}()

		// 继续处理后续的中间件和处理程序
		c.Next()
	}
}

// dumpRequest 格式化请求样式，将 *http.Request 对象转换为字符串形式，方便记录和查看请求信息。
// 参数 req 是一个指向 http.Request 的指针，代表要处理的 HTTP 请求。
// 返回值是一个字符串，包含格式化后的请求信息。
func dumpRequest(req *http.Request) string {
	// 创建一个ReadCloser类型的变量dup，用于后续恢复原始请求体
	var dup io.ReadCloser
	// 将原始Body复制一份给dup，并将req的Body设置为复制的Body，以便后续操作不影响原始请求
	req.Body, dup = dupReadCloser(req.Body)

	// 创建一个bytes.Buffer用于存储格式化后的请求信息
	var b bytes.Buffer
	var err error

	// 获取请求URI
	reqURI := req.RequestURI
	// 如果reqURI为空，则从URL获取RequestURI
	if reqURI == "" {
		reqURI = req.URL.RequestURI()
	}

	// 格式化请求行（方法、URI、HTTP版本）并写入到Buffer中
	_, _ = fmt.Fprintf(&b, "%s %s HTTP/%d.%d\n", req.Method, reqURI, req.ProtoMajor, req.ProtoMinor)
	// 判断请求是否使用了分块传输编码
	chunked := len(req.TransferEncoding) > 0 && req.TransferEncoding[0] == "chunked"
	// 如果请求体不为空
	if req.Body != nil {
		var n int64
		var dest io.Writer = &b
		// 如果使用了分块传输编码，则创建一个新的ChunkedWriter来处理分块数据
		if chunked {
			dest = httputil.NewChunkedWriter(dest)
		}
		// 将请求体复制到Buffer中
		n, err = io.Copy(dest, req.Body)
		// 如果使用了分块传输编码，则关闭ChunkedWriter以确保数据正确写入
		if chunked {
			dest.(io.Closer).Close()
		}
		// 如果复制的数据量大于0，则在Buffer后添加一个换行符，使输出格式更清晰
		if n > 0 {
			_, _ = io.WriteString(&b, "\n")
		}
	}

	// 将req的Body恢复为原始的Body，避免影响后续对该请求的处理
	req.Body = dup
	// 如果复制过程中发生错误，则返回错误信息
	if err != nil {
		return err.Error()
	}

	// 返回Buffer中的内容，即格式化后的请求信息
	return b.String()
}

func dupReadCloser(reader io.ReadCloser) (io.ReadCloser, io.ReadCloser) {
	var buf bytes.Buffer
	tee := io.TeeReader(reader, &buf)
	return io.NopCloser(tee), io.NopCloser(&buf)
}

// dumpStack returns a nicely formatted stack frame, skipping skip frames.
func dumpStack(skip int) []byte {
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
		_, _ = fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}

		_, _ = fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
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
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
