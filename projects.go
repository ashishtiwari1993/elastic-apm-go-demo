package main

import(
    "github.com/gin-gonic/gin"
	"go.elastic.co/apm/module/apmgin/v2"
    "go.elastic.co/apm/v2"
    "net/http"
    "go.elastic.co/apm/module/apmsql/v2"
    _ "go.elastic.co/apm/module/apmsql/v2/mysql"
    "database/sql"
    "github.com/go-redis/redis/v8"
	apmgoredis "go.elastic.co/apm/module/apmgoredisv8/v2"
    "encoding/json"
)

var db *sql.DB
var redisClient *redis.Client

func init() {
    d, err := apmsql.Open("mysql", "root:12345678@tcp(127.0.0.1:3306)/music")
    if err != nil {
        panic(err.Error())
    }

    db = d

    redisClient = redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })
}

func main() {

    router := gin.Default()
    router.Use(apmgin.Middleware(router))

    router.POST("/srvprojects", postProject)
    router.GET("/srvprojects", getProject)
    router.Run("localhost:8003")
}

type project struct {
    UserId string `json:"user_id"`
    Project string `json:"project"`
}

func getProject(c *gin.Context) {

    ctx := c.Request.Context()

    sql := "select user_id, project from projects"
    rows, err := db.QueryContext(ctx, sql)
    checkErr(c,err)

    var projects[] project
    for rows.Next() {
		var p project
		rows.Scan(&p.UserId, &p.Project)
		projects = append(projects, p)
	}

    c.IndentedJSON(http.StatusOK, gin.H{
        "projects": projects,
    })
}

func postProject(c *gin.Context) {

    var newProject project

    err := c.BindJSON(&newProject);
    checkErr(c,err)

    ctx := c.Request.Context()

    sql := "insert into projects values (0,'"+newProject.UserId+"','"+newProject.Project+"')"
    _ , err = db.ExecContext(ctx, sql)
    checkErr(c,err)

    enqueue(c, newProject)

    c.IndentedJSON(http.StatusCreated, newProject)

}

func enqueue(c *gin.Context, p project) {

    ctx := c.Request.Context()
    redisClient.AddHook(apmgoredis.NewHook())
 
    s, err := json.Marshal(p)
    checkErr(c, err)

    _, err = redisClient.Do(ctx, "lpush", "queue1", string(s)).Result()

    checkErr(c, err)
}

func checkErr(c *gin.Context, err error) {

    if err != nil {
        e := apm.CaptureError(c.Request.Context(), err)
        e.Send()
    }
}
