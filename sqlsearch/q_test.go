package sqlsearch

import (
	"testing"
	"time"
)

func TestQ2Sql(t *testing.T) {
	var local, _ = time.LoadLocation("Asia/Shanghai")


	var getKey = func(key string) string {
		return key
	}

	var q1 = `{
	"name": "李南希",
	"desc": "/南希/",
	"age":1,
	"time": "2010-05-18T00:00:00+08:00"
   }`

   var result1 = `name = '李南希' AND desc LIKE '%南希%' AND age = 1 AND time = '2010-05-18 00:00:00'`

   if result := Q2Sql(q1,local,getKey); result != result1 {
   	   t.Log("result is", result, " but expect",result1)
	   t.Fail()
   }

	var q2 = `{
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
}`


	var result2 = `name = '李南希' AND desc LIKE '%南希%' AND created_at >= '2010-05-18 00:00:00' AND created_at < '2018-09-18 19:00:00' AND index > 0 AND index < 99 AND index <> 3 AND index2 = 1 AND test IS NULL AND test2 IS NOT NULL AND tags IN ('appliances','school') AND baby NOT IN ('ok','not_ok') AND age IN (1,2,3)`

	if result := Q2Sql(q2,local,getKey); result != result2 {
		t.Log("result is", result, " but expect",result2, len(result), len(result2),)
		t.Fail()
	}






}
