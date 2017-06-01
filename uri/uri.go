package uri

import (
	"strings"
	"net/url"
	"fmt"
	"errors"
	"strconv"
)


//Uri
type Uri struct {
	Scheme string
	Hosts  hosts
	Username string
	Password string
	Path string
	Query map[string]queryValue
}

type queryValue string


//Parse conver any url to uri
//format：type://[username:password@]host1[:port1][,host2[:port2],...[,hostN[:portN]]][/[database][?options]]
//referecnce:https://docs.mongodb.com/manual/reference/connection-string/
//exp1：mongodb://username:password@host1:8080,host2:8082,host3:8085/database?options1=on&options2=true
//exp2：redis://:password@host1:8080,host2:8082,host3:8085?options1=on&options2=true

func Parse(s string) (*Uri, error) {
	info := &Uri{Query: make(map[string]queryValue)}

	if c := strings.Index(s, "://"); c != -1 {
		info.Scheme = s[:c]
		s = s[c+3:]
	}

	if c := strings.Index(s, "?"); c != -1 {
		for _, pair := range strings.FieldsFunc(s[c+1:], isOptSep) {
			l := strings.SplitN(pair, "=", 2)
			if len(l) != 2 || l[0] == "" || l[1] == "" {
				return nil, errors.New("connection option must be key=value: " + pair)
			}
			info.Query[l[0]] = queryValue(l[1])
		}
		s = s[:c]
	}
	if c := strings.Index(s, "@"); c != -1 {
		pair := strings.SplitN(s[:c], ":", 2)
		if len(pair) > 2 {
			return nil, errors.New("credentials must be provided as user:pass@host")
		}
		var err error
		info.Username, err = url.QueryUnescape(pair[0])
		if err != nil {
			return nil, fmt.Errorf("cannot unescape username in URL: %q", pair[0])
		}
		if len(pair) > 1 {
			info.Password, err = url.QueryUnescape(pair[1])
			if err != nil {
				return nil, fmt.Errorf("cannot unescape password in URL")
			}
		}
		s = s[c+1:]
	}
	if c := strings.Index(s, "/"); c != -1 {
		info.Path = s[c:]
		s = s[:c]
	}
	info.Hosts = strings.Split(s, ",")
	return info, nil
}


func MustParse(s string) *Uri {
	uri, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return uri
}


type hosts []string

func (h *hosts) String() string {
	return strings.Join(*h, ",");
}

//String covert uri to string
func (u *Uri) String() string {
	r := u.NoSchemeString()
	if u.Scheme != "" {
		r = u.Scheme + "://" + r
	}
	return r
}



//QueryRGet remove key and return the value
func (u *Uri) QueryRGet(key string) *queryValue {
	if val, ok := u.Query[key]; ok {
		delete(u.Query, key)
		return &val

	} else {
		return  nil
	}
}

func (u *Uri) NoSchemeString() string {
	r := ""
	var hasUser bool
	if u.Username != "" {
		hasUser = true
		r += u.Username + ":"
	}
	if u.Password != "" {
		if !hasUser {
			r +=  ":"
			hasUser = true
		}
		r += u.Password
	}

	if len(u.Hosts) > 0 {
		if hasUser {
			r += "@"
		}
		r += strings.Join(u.Hosts, ",")
	}
	if u.Path != "" {
		r += u.Path
	}
	if l := len(u.Query); l > 0 {
		var keys []string
		for k, v := range u.Query {
			key := k + "=" + v.String()
			keys = append(keys, key)
		}
		r += "?" + strings.Join(keys, "&")
	}
	return r
}

func isOptSep(c rune) bool {
	return c == ';' || c == '&'
}

func (v *queryValue) String() string {
	return string(*v)
}

func (v *queryValue) MustInt() int {
	i, err := strconv.Atoi(v.String())
	if err != nil {
		panic(err)
	}
	return i
}

func (v *queryValue) MustBool() bool {
	b, err := strconv.ParseBool(v.String())
	if err != nil {
		panic(err)
	}
	return b
}

func (v *queryValue) MustFloat(bitSize int) float64 {
	f, err := strconv.ParseFloat(v.String(), bitSize)
	if err != nil {
		panic(err)
	}
	return f
}
