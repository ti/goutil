# cache map

As explained [here](http://golang.org/doc/faq#atomic_maps) and [here](http://blog.golang.org/go-maps-in-action), the `map` type in Go doesn't support concurrent reads and writes. `concurrent-map` provides a high-performance solution to this by sharding the map with minimal time spent waiting for locks.

## usage


The package is now imported under the "ccmap" namespace. 

## example

```go

	// Create a new map.
	map := cmap.New()
	
	// Sets item within map, sets "bar" under key "foo"
	map.Set("foo", "bar")

	// Retrieve item from map.
	if tmp, ok := map.Get("foo").(string) ok {
		bar := tmp
	}

	// Removes item under key "foo"
	map.Remove("foo")
	
	//ttl
	
	map.SetTTL("foo", "bar", time.Second)
	

```

For more examples have a look at concurrent_map_test.go.

Running tests:

## guidelines for contributing

Contributions are highly welcome. In order for a contribution to be merged, please follow these guidelines:
- Open an issue and describe what you are after (fixing a bug, adding an enhancement, etc.).
- According to the core team's feedback on the above mentioned issue, submit a pull request, describing the changes and linking to the issue.
- New code must have test coverage.
- If the code is about performance issues, you must include benchmarks in the process (either in the issue or in the PR).
- In general, we would like to keep `concurrent-map` as simple as possible and as similar to the native `map`. Please keep this in mind when opening issues.

## license 
MIT (see [LICENSE](https://github.com/orcaman/concurrent-map/blob/master/LICENSE) file)
