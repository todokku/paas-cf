package context

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gin-gonic/gin/render"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// Context is a wrapper type around
// *gin.Context used to implement the
// methods added by ContextInterface
type Context struct {
	*gin.Context
}

func (c *Context) Params() gin.Params {
	return c.Context.Params
}

func (c *Context) Request() *http.Request {
	return c.Context.Request
}

// This interface is extracted from
// the gin.Context type, and
// modified with the addition of the
// methods to replace the properties
// the underlying type exposes.
//
// It's used by the handlers functions, so
// that we can test them fully; including
// the arguments they pass to void-returning
// methods like HTML.
type ContextInterface interface {
	// Property replacements
	Request() *http.Request
	Params() gin.Params

	// Methods
	Copy() *gin.Context
	HandlerName() string
	HandlerNames() []string
	Handler() gin.HandlerFunc
	FullPath() string
	Next()
	IsAborted() bool
	Abort()
	AbortWithStatus(code int)
	AbortWithStatusJSON(code int, jsonObj interface{})
	AbortWithError(code int, err error) *gin.Error
	Error(err error) *gin.Error
	Set(key string, value interface{})
	Get(key string) (value interface{}, exists bool)
	MustGet(key string) interface{}
	GetString(key string) (s string)
	GetBool(key string) (b bool)
	GetInt(key string) (i int)
	GetInt64(key string) (i64 int64)
	GetFloat64(key string) (f64 float64)
	GetTime(key string) (t time.Time)
	GetDuration(key string) (d time.Duration)
	GetStringSlice(key string) (ss []string)
	GetStringMap(key string) (sm map[string]interface{})
	GetStringMapString(key string) (sms map[string]string)
	GetStringMapStringSlice(key string) (smss map[string][]string)
	Param(key string) string
	Query(key string) string
	DefaultQuery(key, defaultValue string) string
	GetQuery(key string) (string, bool)
	QueryArray(key string) []string
	GetQueryArray(key string) ([]string, bool)
	QueryMap(key string) map[string]string
	GetQueryMap(key string) (map[string]string, bool)
	PostForm(key string) string
	DefaultPostForm(key, defaultValue string) string
	GetPostForm(key string) (string, bool)
	PostFormArray(key string) []string
	GetPostFormArray(key string) ([]string, bool)
	PostFormMap(key string) map[string]string
	GetPostFormMap(key string) (map[string]string, bool)
	FormFile(name string) (*multipart.FileHeader, error)
	MultipartForm() (*multipart.Form, error)
	SaveUploadedFile(file *multipart.FileHeader, dst string) error
	Bind(obj interface{}) error
	BindJSON(obj interface{}) error
	BindXML(obj interface{}) error
	BindQuery(obj interface{}) error
	BindYAML(obj interface{}) error
	BindHeader(obj interface{}) error
	BindUri(obj interface{}) error
	MustBindWith(obj interface{}, b binding.Binding) error
	ShouldBind(obj interface{}) error
	ShouldBindJSON(obj interface{}) error
	ShouldBindXML(obj interface{}) error
	ShouldBindQuery(obj interface{}) error
	ShouldBindYAML(obj interface{}) error
	ShouldBindHeader(obj interface{}) error
	ShouldBindUri(obj interface{}) error
	ShouldBindWith(obj interface{}, b binding.Binding) error
	ShouldBindBodyWith(obj interface{}, bb binding.BindingBody) (err error)
	ClientIP() string
	ContentType() string
	IsWebsocket() bool
	Status(code int)
	Header(key, value string)
	GetHeader(key string) string
	GetRawData() ([]byte, error)
	SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)
	Cookie(name string) (string, error)
	Render(code int, r render.Render)
	HTML(code int, name string, obj interface{})
	IndentedJSON(code int, obj interface{})
	SecureJSON(code int, obj interface{})
	JSONP(code int, obj interface{})
	JSON(code int, obj interface{})
	AsciiJSON(code int, obj interface{})
	PureJSON(code int, obj interface{})
	XML(code int, obj interface{})
	YAML(code int, obj interface{})
	ProtoBuf(code int, obj interface{})
	String(code int, format string, values ...interface{})
	Redirect(code int, location string)
	Data(code int, contentType string, data []byte)
	DataFromReader(code int, contentLength int64, contentType string, reader io.Reader, extraHeaders map[string]string)
	File(filepath string)
	FileAttachment(filepath, filename string)
	SSEvent(name string, message interface{})
	Stream(step func(w io.Writer) bool) bool
	Negotiate(code int, config gin.Negotiate)
	NegotiateFormat(offered ...string) string
	SetAccepted(formats ...string)
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
	BindWith(obj interface{}, b binding.Binding) error
}
