package main

import (
	"log"
	"../../sqlsearch"
	"net/url"
	"time"
)

var js = `{
  "name": "李南希",
  "desc": "/南希/",
  "$or": {
    "name": "李南希",
    "desc": "/南希/",
    "created_at": {
      "$gte": "2010-05-18T00:00:00+08:00",
      "$lt": "2018-09-18T19:00:00+08:00"
    },
    "index": {
      "$gt": 0,
      "$lt": 99,
      "$ne": 3
    },
    "index2": {
      "$eq": 1
    },
    "test": {
      "$exists": false
    },
    "test2": {
      "$exists": true
    },
    "tags": {
      "$in": [
        "appliances",
        "school"
      ]
    },
    "baby": {
      "$nin": [
        "ok",
        "not_ok"
      ]
    },
    "age": {
      "$in": [
        1,
        2,
        3
      ]
    }
  }
}`

func main() {
	uriQuery, err := url.ParseQuery(`page=3&limit=10&sort=["-ctime","name"]&select=["name","email","phone","created_at"]&sort=["-created_at"]'&q=%7B%22name%22%3A%22%E6%9D%8E%E5%8D%97%E5%B8%8C%22%2C%22desc%22%3A%22%2F%E5%8D%97%E5%B8%8C%2F%22%2C%22created_at%22%3A%7B%22%24gte%22%3A%222010-05-18T00%3A00%3A00%2B08%3A00%22%2C%22%24lt%22%3A%222018-09-18T19%3A00%3A00%2B08%3A00%22%7D%2C%22age%22%3A%7B%22%24in%22%3A%5B1%2C2%2C3%5D%7D%2C%22tags%22%3A%7B%22%24in%22%3A%5B%22appliances%22%2C%22school%22%5D%7D%7D`)
	if err != nil {
		panic(err)
	}
	_ = sqlsearch.SqlQuery {
		Where:`name = '李南希' AND desc LIKE '%南希%' AND created_at >= '2010-05-18 00:00:00' AND created_at < '2018-09-18 19:00:00' AND age IN (1,2,3) AND tags IN ('appliances','school')`,
		Select:"name,email,phone,created_at",
		Order: "ctime DESC,name",
		Limit:10,
		Offset:20,
	}

	var local, _ = time.LoadLocation("Asia/Shanghai")


	where := sqlsearch.Q2Sql(js, local, func(key string) string {
		return key
	})

	log.Println(where)


	q := sqlsearch.New(uriQuery,local, func(key string) string {
		return "EZ_DEVICE." + key
	})

	sqlQuery := q.ToSql()


	log.Println("Where", sqlQuery.Where)

	log.Println("Select", sqlQuery.Select)

	log.Println("Order", sqlQuery.Order)


}
