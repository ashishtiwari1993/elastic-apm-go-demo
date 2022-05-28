package main

import(
    "github.com/gin-gonic/gin"
	"go.elastic.co/apm/module/apmgin/v2"
    "go.elastic.co/apm/v2"
    "net/http"
    "go.elastic.co/apm/module/apmsql/v2"
    _ "go.elastic.co/apm/module/apmsql/v2/mysql"
    "database/sql"
)

var db *sql.DB

func init() {
    d, err := apmsql.Open("mysql", "root:12345678@tcp(127.0.0.1:3306)/music")
    if err != nil {
        panic(err.Error())
    }

    db = d
}

func main() {

    router := gin.Default()
    router.Use(apmgin.Middleware(router))

    router.POST("/srvtasks", postTask)
    router.Run("localhost:8002")
}

type task struct {
    UserId string `json:"user_id"`
    ProjectId string `json:"project_id"`
    Task string `json:"task"`
}

func postTask(c *gin.Context) {

    var newTask task

    err := c.BindJSON(&newTask);
    checkErr(c,err)

    ctx := c.Request.Context()

    sql := "insert into tasks values (0,'"+newTask.UserId+"','"+newTask.ProjectId+"','"+ newTask.Task +"')"
    _ , err = db.ExecContext(ctx, sql)
    checkErr(c,err)
    c.IndentedJSON(http.StatusCreated, newTask)

}

func checkErr(c *gin.Context, err error) {

    if err != nil {
        e := apm.CaptureError(c.Request.Context(), err)
        e.Send()
    }
}
