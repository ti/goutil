package jsonmap

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"errors"
	"time"
)


var ErrorTypeNotMatch = errors.New("json value type not match")

func GetInt(originalData map[string]interface{}, path string, separatorArr ...string) (v int, err error) {
	value, err := GetProperty(originalData, path, separatorArr...)
	if err != nil {
		return
	}
	if v, ok := value.(int); ok {
		return v, nil
	} else {
		return strconv.Atoi(fmt.Sprint(v))
	}
}

func GetInt64(originalData map[string]interface{}, path string, separatorArr ...string) (v int64, err error) {
	value, err := GetProperty(originalData, path, separatorArr...)
	if err != nil {
		return
	}
	if v, ok := value.(int64); ok {
		return v, nil
	} else {
		return strconv.ParseInt(fmt.Sprint(v), 10, 64)
	}
}

func GetFloat(originalData map[string]interface{}, path string, separatorArr ...string) (v float64, err error) {
	value, err := GetProperty(originalData, path, separatorArr...)
	if err != nil {
		return
	}
	if v, ok := value.(float64); ok {
		return v, nil
	} else {
		return strconv.ParseFloat(fmt.Sprint(v), 64)
	}
}



func GetString(originalData map[string]interface{}, path string, separatorArr ...string) (v string, err error) {
	value, err := GetProperty(originalData, path, separatorArr...)
	if err != nil {
		return
	}
	if v, ok := value.(string); ok {
		return v, nil
	} else if vi, ok := value.(int); ok {
		return strconv.Itoa(vi), nil
	} else {
		return v, ErrorTypeNotMatch
	}
}

func GetBool(originalData map[string]interface{}, path string, separatorArr ...string) (v bool, err error) {
	value, err := GetProperty(originalData, path, separatorArr...)
	if err != nil {
		return
	}
	if v, ok := value.(bool); ok {
		return v, nil
	} else {
		return v, ErrorTypeNotMatch
	}
	return
}


func GetArray(originalData map[string]interface{}, path string, separatorArr ...string) (v []interface{}, err error) {
	value, err := GetProperty(originalData, path, separatorArr...)
	if err != nil {
		return
	}
	if v, ok := value.([]interface{}); ok {
		return v, nil
	} else {
		return v, ErrorTypeNotMatch
	}
	return
}


func GetMap(originalData map[string]interface{}, path string, separatorArr ...string) (v map[string]interface{}, err error) {
	value, err := GetProperty(originalData, path, separatorArr...)
	if err != nil {
		return
	}
	if v, ok := value.(map[string]interface{}); ok {
		return v, nil
	} else {
		return v, ErrorTypeNotMatch
	}
	return
}



func GetTime(originalData map[string]interface{}, path string, separatorArr ...string) (v time.Time, err error) {
	value, err := GetProperty(originalData, path, separatorArr...)
	if err != nil {
		return
	}
	if v, ok := value.(time.Time); ok {
		return v, nil
	} else if vStr, ok := value.(string); ok{
		if t, errTime := time.Parse(time.RFC3339, vStr); errTime == nil {
			return t, nil
		} else {
			return v, ErrorTypeNotMatch
		}
	} else {
		return v, ErrorTypeNotMatch
	}
	return
}



// GetProperty returns a property if it exist.
//
//    property, err := GetProperty(document, "one.two.three[0]")
//    property, err := GetProperty(document, "one.two.three[0]", ".")
//    property, err := GetProperty(document, "one/two/three[0]", "/")
//
// Property type is `interface{}`
func GetProperty(originalData map[string]interface{}, path string, separatorArr ...string) (pathParsed interface{}, err error) {
	var separator = "."
	if len(separatorArr) > 0 {
		if len(separatorArr[0]) > 0 {
			separator = separatorArr[0]
		}
	}

	data := make(map[string]interface{})
	d := reflect.ValueOf(originalData)
	for _, key := range d.MapKeys() {
		data[key.String()] = d.MapIndex(key).Interface()
	}

	err = fmt.Errorf("Property %s does not exist", path)

	if len(path) == 0 {
		path = separator
	}

	levels_tmp := strings.Split(path, separator)
	levels := make([]string, 0)
	for _, level_tmp := range levels_tmp {
		if len(level_tmp) > 0 {
			levels = append(levels, level_tmp)
		}
	}

	if len(levels) > 0 && path != separator {
		path_level_one := levels[0]
		// If we have a level in path_level_one

		re := regexp.MustCompile(`(\w+[\_]?[\-]?)+\[\d+\]{1}`)
		if matched := re.FindString(path_level_one); len(matched) > 0 {
			property_re := regexp.MustCompile(`(\w+[\_]?[\-]?)+`)
			index_re := regexp.MustCompile(`\[\d+\]{1}`)
			// Get a property
			// avatars
			property := property_re.FindString(path_level_one)

			// Get an index
			index_found := index_re.FindString(path_level_one)

			// If index > 0 - check if this property is array
			if len(index_found) > 0 {
				if len(property) > 0 {
					path_level_one = property
				}
				index_found = strings.Trim(index_found, "[]")
				if index, err := strconv.Atoi(index_found); err == nil {
					if v, ok := data[property]; ok {
						if isKind(v, reflect.Slice) {
							slice := reflect.ValueOf(v)
							if index >= 0 && index < slice.Len() {
								value := slice.Index(index).Interface()

								data[property] = value
							} else {
								err = fmt.Errorf(
									"%s: Min index is 0, Max index is %d. You passed index %d", property, slice.Len(), index,
								)
								return pathParsed, err
							}
						} else {
							err = fmt.Errorf(
								"%s: is not an array", property,
							)
							return pathParsed, err
						}
					} else {
						err = fmt.Errorf(
							"Property %s does not exist", property,
						)
						return pathParsed, err
					}
				} else {
					err = fmt.Errorf(
						"%s must be of type %s",
						fmt.Sprintf("%s[%d]", property, index_found),
						"number",
					)
					return pathParsed, err
				}
			}
		}

		if len(levels[1:]) >= 1 {
			if level_one_value, ok := data[path_level_one]; ok {
				if level_one_value != nil {
					switch reflect.TypeOf(level_one_value).Kind() {
					case reflect.Map:
						mapped_level_one_value := make(map[string]interface{})
						d := reflect.ValueOf(level_one_value)
						for _, key := range d.MapKeys() {
							mapped_level_one_value[key.String()] = d.MapIndex(key).Interface()
						}
						return GetProperty(mapped_level_one_value, strings.Join(levels[1:], separator), separator)
					default:
						// pass
					}
				}
			} else {
				err = fmt.Errorf(
					"Property %s does not exist", path_level_one,
				)
				return pathParsed, err
			}
		} else {
			if v, ok := data[path_level_one]; ok {
				pathParsed = v
				err = nil
			}
		}
	} else if path == separator {
		pathParsed = data
		err = nil
	}
	return
}

// DeleteProperty removes a property from map
//
//    err := GetProperty(document, "one.two.three[0]")
//    err := GetProperty(document, "one.two.three[0]", ".")
//    err := GetProperty(document, "one/two/three[0]", "/")
//
func DeleteProperty(originalData map[string]interface{}, path string, separatorArr ...string) (err error) {
	var separator = "."
	if len(separatorArr) > 0 {
		if len(separatorArr[0]) > 0 {
			separator = separatorArr[0]
		}
	}

	// If we have a property
	if _, err = GetProperty(originalData, path, separator); err != nil {
		return
	}

	if len(path) == 0 {
		path = separator
	}

	levels_tmp := strings.Split(path, separator)
	levels := make([]string, 0)
	for _, level_tmp := range levels_tmp {
		if len(level_tmp) > 0 {
			levels = append(levels, level_tmp)
		}
	}

	if len(levels) > 0 && path != separator {
		path_level_one := levels[0]

		// If we have a level in path_level_one

		re := regexp.MustCompile(`(\w+[\_]?[\-]?)+\[\d+\]{1}`)
		if matched := re.FindString(path_level_one); len(matched) > 0 {
			property_re := regexp.MustCompile(`(\w+[\_]?[\-]?)+`)
			index_re := regexp.MustCompile(`\[\d+\]{1}`)
			// Get a property
			// avatars
			property := property_re.FindString(path_level_one)

			// Get an index
			index_found := index_re.FindString(path_level_one)

			// If index > 0 - check if this property is array
			if len(index_found) > 0 {
				if len(property) > 0 {
					path_level_one = property
				}
				index_found = strings.Trim(index_found, "[]")
				if index, err := strconv.Atoi(index_found); err == nil {
					if v, ok := originalData[property]; ok {
						if isKind(v, reflect.Slice) {
							slice := reflect.ValueOf(v)
							if index >= 0 && index < slice.Len() {
								value := slice.Index(index).Interface()
								// If len of other levels greater than 0
								if len(levels[1:]) >= 1 {
									if isKind(value, reflect.Map) {
										mapped_value := value.(map[string]interface{})
										err = DeleteProperty(mapped_value, strings.Join(levels[1:], separator), separator)
										if err == nil {
											// If we have an empty value inside of a slice - remove it
											if len(mapped_value) == 0 {
												slices := make([]interface{}, 0)
												for i := 0; i < slice.Len(); i++ {
													if i != index {
														slices = append(slices, slice.Index(i).Interface())
													}
												}
												originalData[property] = slices
											}
										}
										return err
									}
								} else {
									// if this is a `property[1]` in a path like `path.to.property[1]`
									slices := make([]interface{}, 0)
									for i := 0; i < slice.Len(); i++ {
										if i != index {
											slices = append(slices, slice.Index(i).Interface())
										}
									}
									originalData[property] = slices
									return err
								}
							} else {
								err = fmt.Errorf(
									"%s: Min index is 0, Max index is %d", property, slice.Len(),
								)
							}
						} else {
							err = fmt.Errorf(
								"%s: is not an array", property,
							)
						}
					} else {
						err = fmt.Errorf(
							"Property %s does not exist", property,
						)
					}
				} else {
					err = fmt.Errorf(
						"%s must be of type %s",
						fmt.Sprintf("%s[%d]", property, index_found),
						"number",
					)
				}
			}
		}
		if err != nil {
			return err
		}

		if len(levels[1:]) >= 1 {
			if level_one_value, ok := originalData[path_level_one]; ok {
				if level_one_value != nil {
					switch reflect.TypeOf(level_one_value).Kind() {
					case reflect.Map:
						if mapped_level_one_value, ok := level_one_value.(map[string]interface{}); ok {
							err = DeleteProperty(mapped_level_one_value, strings.Join(levels[1:], separator), separator)
							if err != nil {
								return
							}
						}
					default:
						delete(originalData, path)
					}

				}
			} else {
				err = fmt.Errorf(
					"Property %s does not exist", path_level_one,
				)
				return
			}
		} else {
			delete(originalData, path_level_one)
		}
	} else if path == separator {
		for k := range originalData {
			delete(originalData, k)
		}
	}

	return
}

// AddProperty adds a property to map. Returns an error if property already exists
//
//    err := AddProperty(document, "one.two.three[0]", "string value")
//    err := AddProperty(document, "one.two.three[0]", "string value", ".")
//    err := AddProperty(document, "one/two/three[0]", "string value", "/")
//
func AddProperty(originalData map[string]interface{}, path string, value interface{}, separatorArr ...string) (err error) {
	var separator = "."
	if len(separatorArr) > 0 {
		if len(separatorArr[0]) > 0 {
			separator = separatorArr[0]
		}
	}

	// If we have a property - raise an error
	if _, err = GetProperty(originalData, path, separator); err == nil {
		err = fmt.Errorf(
			"Property %s already exists", path,
		)
		return
	} else {
		err = nil
	}

	if len(path) == 0 {
		path = separator
	}

	levels_tmp := strings.Split(path, separator)
	levels := make([]string, 0)
	for _, level_tmp := range levels_tmp {
		if len(level_tmp) > 0 {
			levels = append(levels, level_tmp)
		}
	}

	if len(levels) > 0 && path != separator {
		path_level_one := levels[0]
		// If we have a level in path_level_one

		re := regexp.MustCompile(`(\w+[\_]?[\-]?)+\[\d+\]{1}`)
		if matched := re.FindString(path_level_one); len(matched) > 0 {
			property_re := regexp.MustCompile(`(\w+[\_]?[\-]?)+`)
			index_re := regexp.MustCompile(`\[\d+\]{1}`)
			// Get a property
			// avatars
			property := property_re.FindString(path_level_one)

			// Get an index
			index_found := index_re.FindString(path_level_one)

			// If index > 0 - check if this property is array
			if len(index_found) > 0 {
				if len(property) > 0 {
					path_level_one = property
				}
				index_found = strings.Trim(index_found, "[]")
				if index, err := strconv.Atoi(index_found); err == nil {
					if v, ok := originalData[property]; ok {
						if isKind(v, reflect.Slice) {
							slice := reflect.ValueOf(v)
							var dest_value interface{}

							if index < slice.Len() {
								dest_value = slice.Index(index).Interface()
							}

							// If len of other levels greater than 0
							if len(levels[1:]) >= 1 {
								if isKind(dest_value, reflect.Map) {
									mapped_value := dest_value.(map[string]interface{})
									return AddProperty(mapped_value, strings.Join(levels[1:], separator), value, separator)
								} else if dest_value == nil {
									slice_len := slice.Len()
									if index > slice_len-1 {
										slice_len = index + 1
									}
									vv := make([]interface{}, slice_len)

									for i, _ := range vv {
										var val interface{}
										if i < slice.Len() {
											val = slice.Index(i).Interface()
										}
										vv[i] = val
									}

									mapped_value := make(map[string]interface{})

									err = AddProperty(
										mapped_value,
										strings.Join(levels[1:], separator),
										value,
										separator,
									)
									vv[index] = mapped_value
									originalData[property] = vv

									return err
								}
							} else {
								// if this is a `property[1]` in a path like `path.to.property[1]`
								slice_len := slice.Len()
								if index > slice_len-1 {
									slice_len = index + 1
								}
								slices := make([]interface{}, slice_len)

								for i := 0; i < slice.Len(); i++ {
									if i < slice.Len() {
										slices[i] = slice.Index(i).Interface()
									}
								}
								slices[index] = value

								originalData[path_level_one] = slices
								return err
							}
						} else {
							err = fmt.Errorf(
								"%s: is not an array", property,
							)
						}
					} else {
						new_sliced_value := make([]interface{}, index+1)
						new_mapped_value := make(map[string]interface{})
						new_sliced_value[index] = new_mapped_value
						originalData[path_level_one] = new_sliced_value
						return UpdateProperty(originalData, path, value, separator)
					}
				} else {
					err = fmt.Errorf(
						"%s must be of type %s",
						fmt.Sprintf("%s[%d]", property, index_found),
						"number",
					)
				}
			}
		}

		if err != nil {
			return err
		}

		if len(levels[1:]) >= 1 {
			if level_one_value, ok := originalData[path_level_one]; ok {
				if level_one_value != nil {
					switch reflect.TypeOf(level_one_value).Kind() {
					case reflect.Map:
						if mapped_level_one_value, ok := level_one_value.(map[string]interface{}); ok {
							return AddProperty(mapped_level_one_value, strings.Join(levels[1:], separator), value, separator)
						}
					default:
						originalData[path] = value
					}

				}
			} else {
				new_mapped_value := make(map[string]interface{})
				originalData[path_level_one] = new_mapped_value
				return AddProperty(new_mapped_value, strings.Join(levels[1:], separator), value, separator)
			}
		} else {
			// If a map does not contain a last node property
			if _, ok := originalData[path_level_one]; !ok {
				originalData[path_level_one] = value
			}
		}
	} else if path == separator {
		originalData[path] = value
	}
	return
}

// UpdateProperty updates a property in a map. It will create or update existing property
//
//    err := UpdateProperty(document, "one.two.three[0]", "string value")
//    err := UpdateProperty(document, "one.two.three[0]", "string value", ".")
//    err := UpdateProperty(document, "one/two/three[0]", "string value", "/")
//
func UpdateProperty(originalData map[string]interface{}, path string, value interface{}, separatorArr ...string) (err error) {
	var separator = "."
	if len(separatorArr) > 0 {
		if len(separatorArr[0]) > 0 {
			separator = separatorArr[0]
		}
	}

	// If we have a property - update it, otherwise add it
	if _, err = GetProperty(originalData, path, separator); err != nil {
		return AddProperty(originalData, path, value, separator)
	} else {
		if len(path) == 0 {
			path = separator
		}

		levels_tmp := strings.Split(path, separator)
		levels := make([]string, 0)
		for _, level_tmp := range levels_tmp {
			if len(level_tmp) > 0 {
				levels = append(levels, level_tmp)
			}
		}

		if len(levels) > 0 && path != separator {
			path_level_one := levels[0]

			// If we have a level in path_level_one
			re := regexp.MustCompile(`(\w+[\_]?[\-]?)+\[\d+\]{1}`)
			if matched := re.FindString(path_level_one); len(matched) > 0 {
				property_re := regexp.MustCompile(`(\w+[\_]?[\-]?)+`)
				index_re := regexp.MustCompile(`\[\d+\]{1}`)
				property := property_re.FindString(path_level_one)

				// Get an index
				index_found := index_re.FindString(path_level_one)

				// If index > 0 - check if this property is array
				if len(index_found) >= 0 {
					if len(property) > 0 {
						path_level_one = property
					}
					index_found = strings.Trim(index_found, "[]")
					if index, err := strconv.Atoi(index_found); err == nil {
						if v, ok := originalData[property]; ok {
							if isKind(v, reflect.Slice) {
								slice := reflect.ValueOf(v)
								var dest_value interface{}
								if index >= 0 && index < slice.Len() {
									dest_value = slice.Index(index).Interface()
								}
								// If len of other levels greater than 0
								if len(levels[1:]) >= 1 {
									if isKind(dest_value, reflect.Map) {
										mapped_value := dest_value.(map[string]interface{})
										return UpdateProperty(mapped_value, strings.Join(levels[1:], separator), value, separator)
									}
								} else {
									// if this is a `property[1]` in a path like `path.to.property[1]`
									slices := make([]interface{}, slice.Len())
									for i := 0; i < slice.Len(); i++ {
										if i != index {
											slices[i] = slice.Index(i).Interface()
										} else {
											slices[index] = value
										}
									}
									if index >= slice.Len() {
										slices = append(slices, value)
									}

									originalData[path_level_one] = slices
									return err
								}
							}
						} else {
							err = fmt.Errorf(
								"%s: is not an array", property,
							)
						}
					} else {
						err = fmt.Errorf(
							"Property %s does not exist", property,
						)
					}
				} else {
					err = fmt.Errorf(
						"%s must be of type %s",
						fmt.Sprintf("%s[%d]", property, index_found),
						"number",
					)
				}
			}

			if len(levels[1:]) >= 1 {
				if level_one_value, ok := originalData[path_level_one]; ok {
					if level_one_value != nil {
						switch reflect.TypeOf(level_one_value).Kind() {
						case reflect.Map:
							if mapped_level_one_value, ok := level_one_value.(map[string]interface{}); ok {
								return UpdateProperty(mapped_level_one_value, strings.Join(levels[1:], separator), value, separator)
							}
						default:
							originalData[path] = value
						}
					}
				} else {
					err = fmt.Errorf(
						"Property %s does not exist", path_level_one,
					)
					return
				}
			} else {
				originalData[path_level_one] = value
			}
		} else if path == separator {
			originalData[path] = value
		}
	}

	return
}
