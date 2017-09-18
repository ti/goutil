package sqlsearch

import (
	"testing"
	"net/url"
	"log"
	"time"
)

func TestUrlQuery2Sql(t *testing.T) {
	expect := SqlQuery {
		Where:`name = '李南希' AND desc LIKE '%南希%' AND created_at >= '2010-05-18 00:00:00' AND created_at < '2018-09-18 19:00:00' AND age IN (1,2,3) AND tags IN ('appliances','school')`,
		Select:"name,email,phone,created_at",
		Order: "ctime DESC,name",
		Limit:10,
		Offset:20,
	}

	uriQuery, err := url.ParseQuery(`page=3&limit=10&sort=["-ctime","name"]&select=["name","email","phone","created_at"]&sort=["-created_at"]'&q=%7B%22name%22%3A%22%E6%9D%8E%E5%8D%97%E5%B8%8C%22%2C%22desc%22%3A%22%2F%E5%8D%97%E5%B8%8C%2F%22%2C%22created_at%22%3A%7B%22%24gte%22%3A%222010-05-18T00%3A00%3A00%2B08%3A00%22%2C%22%24lt%22%3A%222018-09-18T19%3A00%3A00%2B08%3A00%22%7D%2C%22age%22%3A%7B%22%24in%22%3A%5B1%2C2%2C3%5D%7D%2C%22tags%22%3A%7B%22%24in%22%3A%5B%22appliances%22%2C%22school%22%5D%7D%7D`)
	if err != nil {
		panic(err)
	}

	var local, _ = time.LoadLocation("Asia/Shanghai")

	q := New(uriQuery,local, nil)


	sqlQuery := q.ToSql()

	if *sqlQuery != expect {
		log.Println(sqlQuery.Where)
		t.Fail()
	}

}
