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

    router.POST("/srvusers", postUser)
    router.Run("localhost:8001")
}

type user struct {
    Name string `json:"name"`
    Dept string `json:"dept"`
    Role string `json:"role"`
}

func postUser(c *gin.Context) {

    var newUser user

    err := c.BindJSON(&newUser);
    checkErr(c,err)

    ctx := c.Request.Context()

    sql := "insert into users values (0,'"+newUser.Name+"','"+newUser.Dept+"','"+ newUser.Role +"')"
    _ , err = db.ExecContext(ctx, sql)
    checkErr(c,err)

    c.IndentedJSON(http.StatusCreated, newUser)

}

func checkErr(c *gin.Context, err error) {

    if err != nil {
        e := apm.CaptureError(c.Request.Context(), err)
        e.Send()
    }
}
