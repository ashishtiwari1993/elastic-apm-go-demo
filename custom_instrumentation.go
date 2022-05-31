package main

import (
    "go.elastic.co/apm/v2"
    "errors"
)

func main() {

    //start transaction
    tx := apm.DefaultTracer().StartTransaction("message-in-topic-consumer", "request")

    //start span
    span := tx.StartSpan("SELECT FROM foo", "db.mysql.query", nil)
    span.End()

    //Push Error
    err := errors.New("Panic error:testing")
    e := apm.DefaultTracer().NewError(err)
    e.SetTransaction(tx)
    e.Send()

    //transaction end
    tx.Result = "error"
    tx.End()

    apm.DefaultTracer().Flush(nil)
}
