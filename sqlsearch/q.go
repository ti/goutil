package sqlsearch

import (
	"strings"
	"time"
	"strconv"
	"log"
	"encoding/json"
)

var mg2SqlGroup = map[string]struct {
	Type int
	Sql  string
}{
	"$eq":     {0, "="},
	"$ne":     {0, "<>"},
	"$gt":     {0, ">"},
	"$gte":    {0, ">="},
	"$lt":     {0, "<"},
	"$lte":    {0, "<="},
	"$in":     {1, "IN"},
	"$nin":    {1, "NOT IN"},
	"$exists": {2, "IS NOT NULL"},
}

func Q2Sql(queryStr string, timeLocal *time.Location, getKey func(k string) string) string {
	var query OrderedMap
	if err := json.Unmarshal([]byte(queryStr), &query); err != nil {
		return ""
	}
	var wheres []string
	for _, key := range query.Keys() {
		value := query.MustGet(key)
		if stringValue, ok := value.(string); ok {
			if strings.HasPrefix(stringValue, "/") && strings.HasSuffix(stringValue, "/") {
				v := stringValue[1:len(stringValue)-1]
				wheres = append(wheres, getKey(key)+" LIKE '%"+v+"%'")
			} else {
				if tm, err := time.Parse(time.RFC3339, value.(string)); err == nil {
					wheres = append(wheres, getKey(key)+" = '"+tm.In(timeLocal).Format("2006-01-02 15:04:05")+"'")
				} else {
					wheres = append(wheres, getKey(key)+" = '"+stringValue+"'")
				}
			}
		} else if mapValue, ok := value.(OrderedMap); ok {
			for _, k := range mapValue.Keys() {
				v := mapValue.MustGet(k)
				//凡是以"$"开头是字段筛选条件，添加对日期取值的操作
				if strings.HasPrefix(k, "$") {
					grp := mg2SqlGroup[k]
					switch grp.Type {
					case 0:
						conValue := toSqlString(v, timeLocal)
						if conValue != "" {
							wheres = append(wheres, getKey(key)+" "+grp.Sql+" "+conValue)
						}
					case 1:
						var inValue string
						if slices, ok := v.([]interface{}); ok && len(slices) > 0 {
							for i, v := range slices {
								if i > 0 {
									inValue += ","
								}
								inValue += toSqlString(v, nil)
							}
						}
						if inValue != "" {
							wheres = append(wheres, getKey(key)+" "+grp.Sql+" ("+inValue+")")
						}
					case 2:
						if k == "$exists" {
							if exists, ok := v.(bool); ok {
								if exists {
									wheres = append(wheres, getKey(key)+" "+grp.Sql)
								} else {
									wheres = append(wheres, getKey(key)+" IS NULL")
								}
							}
						}
					default:
						//DO NOTING
					}
				}
			}
		} else {
			wheres = append(wheres, getKey(key)+" = "+toSqlString(value, nil))
		}
	}
	return strings.Join(wheres, " AND ")
}

func toSqlString(v interface{}, timeLocal *time.Location) (result string) {
	switch t := v.(type) {
	case int:
		result = strconv.Itoa(v.(int))
	case int64:
		result = strconv.FormatInt(v.(int64), 10)
	case float64:
		result = strconv.FormatFloat(v.(float64), 'f', -1, 64)
	case string:
		if timeLocal != nil {
			if tm, err := time.Parse(time.RFC3339, v.(string)); err == nil {
				result = "'" + tm.In(timeLocal).Format("2006-01-02 15:04:05") + "'"
			} else {
				result = "'" + v.(string) + "'"
			}
		} else {
			result = "'" + v.(string) + "'"
		}
	default:
		log.Println("toSqlString unkown type ", t)
	}
	return
}
