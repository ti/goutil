package uri

import (
	"testing"
)


func TestUriParse(t *testing.T) {
	uris := []string{
		"mongodb://username:password@host1:8080,host2:8082,host3:8085/database?options1=on&options2=true",
		"redis://:password@host1:8080,host2:8082,host3:8085",
		"lcoalhost",
		"data.sp.cn",
	}

	exps := []Uri{
		Uri{Scheme: "mongodb", Username: "username", Password: "password", Hosts: []string{"host1:8080", "host2:8082", "host3:8085", }, Query: map[string]queryValue{"options1": "on", "options2": "true"}},
		Uri{Scheme: "redis", Password: "password", Hosts: []string{"host1:8080", "host2:8082", "host3:8085", }},
		Uri{Hosts: []string{"lcoalhost"}},
		Uri{Hosts: []string{"data.sp.cn"}},
	}

	for i:=0; i < len(uris); i++ {
		uri, err := Parse(uris[i])
		if err != nil {
			t.Fatal("parse error", uris[i], err)
		}
		exp := exps[i]

		if exp.Scheme != uri.Scheme || exp.Username != uri.Username || exp.Password != uri.Password{
			t.Fatal(uris[i], "does not match expected")
		}
		for j:=0; j < len(uri.Hosts); j++ {
			if uri.Hosts[j] != exp.Hosts[j] {
				t.Fatal(uris[i], "'s hosts does not match expected")
			}
		}

		if len(uri.Query) != len(exp.Query) {
			t.Fatal(uris[i], "'s query does not match expected")
		}
		for k, v := range exp.Query {
			if uri.Query[k] != v {
				t.Fatal(uris[i], "'s query does not match expected")
			}
		}

	}

}



func TestUri_QueryRGet(t *testing.T) {
	url := "mongodb://username:password@host1:8080,host2:8082,host3:8085/database?options1=on&options2=true"
	uri, err := Parse(url)
	if err != nil {
		t.Fatal("parse error", url, err)
	}
	opt2 := uri.QueryRGet("options2").MustBool()
	if opt2  != true  || uri.String() != "mongodb://username:password@host1:8080,host2:8082,host3:8085/database?options1=on" {
		t.Fatal("parse error", url, err)
	}
}