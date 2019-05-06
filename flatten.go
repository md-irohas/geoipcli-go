// This file is a modified code of the following repository.
// https://github.com/doublerebel/bellows/blob/master/main.go
//
// The original copyright is as follows:

// Copyright © 2016 Charles Phillips <charles@doublerebel.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
)

func flatten(value interface{}, prefix string, m map[string]interface{}) {
	base := ""
	if prefix != "" {
		base = prefix + "."
	}

	orig := reflect.ValueOf(value)
	kind := orig.Kind()
	if kind == reflect.Ptr || kind == reflect.Interface {
		orig = reflect.Indirect(orig)
		kind = orig.Kind()
	}

	t := orig.Type()

	switch kind {
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			break
		}

		keys := orig.MapKeys()
		if len(keys) == 0 {
			flatten("", base+"[language]", m)
		} else {
			for _, childKey := range keys {
				childValue := orig.MapIndex(childKey)
				flatten(childValue.Interface(), base+childKey.String(), m)
			}
		}
	case reflect.Struct:
		for i := 0; i < orig.NumField(); i += 1 {
			childValue := orig.Field(i)
			childKey := t.Field(i).Tag.Get("maxminddb")
			flatten(childValue.Interface(), base+childKey, m)
		}
	default:
		if prefix != "" {
			m[prefix] = value
		}
	}
}

func Flatten(value interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	flatten(value, "", m)
	return m
}
