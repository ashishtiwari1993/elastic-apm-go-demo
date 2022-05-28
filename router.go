package main

import(
    "github.com/gin-gonic/gin"
	"go.elastic.co/apm/module/apmgin/v2"
    "go.elastic.co/apm/module/apmhttp/v2"
    "go.elastic.co/apm/v2"
    "net/url"
    "net/http"
    "net/http/httputil"
    "github.com/go-redis/redis/v8"
	apmgoredis "go.elastic.co/apm/module/apmgoredisv8/v2"
//    "fmt"
)

var redisClient *redis.Client

func init() {
    redisClient = redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })
}

func main() {

    router := gin.Default()
    router.Use(apmgin.Middleware(router))

    router.Any("/users", checkAuth, usersProxy)
    router.Any("/tasks", checkAuth, tasksProxy)
    router.Run("localhost:8000")
}

func checkAuth(c *gin.Context) {

    ctx := c.Request.Context()
    redisClient.AddHook(apmgoredis.NewHook())

    api_key := c.Query("apikey")

    user_key, err := redisClient.Get(ctx, "apikey").Result()

    if err == redis.Nil {
        c.Abort()
        c.JSON(http.StatusUnauthorized, gin.H{"Error": "Please contact"})
        return
    } else if err != nil {

        checkErr(c,err)
        c.Abort()
        c.JSON(http.StatusUnauthorized, gin.H{"Error": "Please contact"})
        return
    } else if api_key != user_key {

        c.Abort()
        c.JSON(http.StatusUnauthorized, gin.H{"Error": "Invalid API key."})
        return
    }

    c.Next()
}

func usersProxy(c *gin.Context) {

    remote, err := url.Parse("http://localhost:8001/srvusers")
	checkErr(c,err)

    //extracting traceparent and tracestate
    tx := apm.TransactionFromContext(c.Request.Context())
    traceContext := tx.TraceContext()
    traceparent := apmhttp.FormatTraceparentHeader(traceContext) 
    tracestate := traceContext.State.String()

    c.Request.Header.Add("Traceparent", traceparent)
    c.Request.Header.Add("Tracestate", tracestate)

	proxy := httputil.NewSingleHostReverseProxy(remote)

	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = remote.Path
	}
	proxy.ServeHTTP(c.Writer, c.Request)    
}

func tasksProxy(c *gin.Context) {

    remote, err := url.Parse("http://localhost:8002/srvtasks")
	checkErr(c,err)

    //extracting traceparent and tracestate
    tx := apm.TransactionFromContext(c.Request.Context())
    traceContext := tx.TraceContext()
    traceparent := apmhttp.FormatTraceparentHeader(traceContext) 
    tracestate := traceContext.State.String()

    c.Request.Header.Add("Traceparent", traceparent)
    c.Request.Header.Add("Tracestate", tracestate)

	proxy := httputil.NewSingleHostReverseProxy(remote)

	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = remote.Path
	}
	proxy.ServeHTTP(c.Writer, c.Request)    
}


func checkErr(c *gin.Context, err error) {

    if err != nil {
        e := apm.CaptureError(c.Request.Context(), err)
        e.Send()
    }
}
