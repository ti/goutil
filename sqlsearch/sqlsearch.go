package sqlsearch

import (
	"net/url"
	"strconv"
	"encoding/json"
	"strings"
	"time"
)


type SqlQuery  struct {
	Where string
	Select string
	Order  string
	Offset int
	Limit int
}


type SqlSearch struct {
	UrlValues    url.Values
	TimeLocation *time.Location
	KeyAlias   func(key string)string
}

func New(query url.Values, local *time.Location, keyAlias func(key string)string) *SqlSearch {
	ret := SqlSearch{
		UrlValues:query,
		TimeLocation: local,
	}
	if keyAlias == nil {
		ret.KeyAlias = func(key string) string {
			return key
		}
	} else {
		ret.KeyAlias = keyAlias
	}
	return &ret
}

func (this *SqlSearch) ToSql() *SqlQuery {
	sq := SqlQuery{Limit: 2000,}
	if lmt := this.UrlValues.Get("limit"); lmt != "" {
		if lmtInt, err := strconv.Atoi(lmt); err == nil && lmtInt < 2000 {
			sq.Limit = lmtInt
		}
	}
	if qSelect := this.UrlValues.Get("select"); qSelect != "" {
		var searchSelect []string
		if err := json.Unmarshal([]byte(qSelect), &searchSelect); err == nil {
			selectStr := ""
			for i, v := range searchSelect {
				if i > 0 {
					selectStr += ","
				}
				if strings.HasPrefix(v, "-") {
					selectStr = ""
					continue
				}
				selectStr += this.KeyAlias(v)
			}
			sq.Select = selectStr
		}
	}

	if qSort := this.UrlValues.Get("sort"); qSort != "" {
		var searchSort []string
		if err := json.Unmarshal([]byte(qSort), &searchSort); err == nil {
			sort := ""
			for i, v := range searchSort {
				if i > 0 {
					sort += ","
				}
				if strings.HasPrefix(v, "-") {
					sort += this.KeyAlias(v[1:]) + " DESC"
				} else {
					sort += this.KeyAlias(v)
				}
			}
			sq.Order = sort
		}
	}

	if page := this.UrlValues.Get("page"); page != "" {
		if pageInt, err := strconv.Atoi(page);err == nil {
			sq.Offset = sq.Limit * (pageInt - 1)
		}
	}
	if q := this.UrlValues.Get("q"); q != "" {
		sq.Where = Q2Sql(q, this.TimeLocation, this.KeyAlias)
	}
	return &sq
}