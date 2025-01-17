package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultDecoder(t *testing.T) {
	src := &KeyValue{
		Key:    "service",
		Value:  []byte("config"),
		Format: "",
	}
	target := make(map[string]interface{}, 0)
	err := defaultDecoder(src, target)
	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{
		"service": []byte("config"),
	}, target)

	src = &KeyValue{
		Key:    "service.name.alias",
		Value:  []byte("2233"),
		Format: "",
	}
	target = make(map[string]interface{}, 0)
	err = defaultDecoder(src, target)
	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{
		"service": map[string]interface{}{
			"name": map[string]interface{}{
				"alias": []byte("2233"),
			},
		},
	}, target)
}

func TestDefaultResolver(t *testing.T) {
	var (
		portString = "8080"
		countInt   = 10
		rateFloat  = 0.9
	)

	data := map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": map[string]interface{}{
				"notexist": "${NOTEXIST:100}",
				"port":     "${PORT:8081}",
				"count":    "${COUNT:0}",
				"enable":   "${ENABLE:false}",
				"rate":     "${RATE}",
				"empty":    "${EMPTY:foobar}",
				"url":      "${URL:http://example.com}",
				"array": []interface{}{
					"${PORT}",
					map[string]interface{}{"foobar": "${NOTEXIST:8081}"},
				},
				"value1": "${test.value}",
				"value2": "$PORT",
				"value3": "abc${PORT}foo${COUNT}bar",
				"value4": "${foo${bar}}",
			},
		},
		"test": map[string]interface{}{
			"value": "foobar",
		},
		"PORT":   "8080",
		"COUNT":  "10",
		"ENABLE": "true",
		"RATE":   "0.9",
		"EMPTY":  "",
	}

	tests := []struct {
		name   string
		path   string
		expect interface{}
	}{
		{
			name:   "test not exist int env with default",
			path:   "foo.bar.notexist",
			expect: 100,
		},
		{
			name:   "test string with default",
			path:   "foo.bar.port",
			expect: portString,
		},
		{
			name:   "test int with default",
			path:   "foo.bar.count",
			expect: countInt,
		},
		{
			name:   "test bool with default",
			path:   "foo.bar.enable",
			expect: true,
		},
		{
			name:   "test float without default",
			path:   "foo.bar.rate",
			expect: rateFloat,
		},
		{
			name:   "test empty value with default",
			path:   "foo.bar.empty",
			expect: "",
		},
		{
			name:   "test url with default",
			path:   "foo.bar.url",
			expect: "http://example.com",
		},
		{
			name:   "test array",
			path:   "foo.bar.array",
			expect: []interface{}{portString, map[string]interface{}{"foobar": "8081"}},
		},
		{
			name:   "test ${test.value}",
			path:   "foo.bar.value1",
			expect: "foobar",
		},
		{
			name:   "test $PORT",
			path:   "foo.bar.value2",
			expect: "$PORT",
		},
		{
			name:   "test abc${PORT}foo${COUNT}bar",
			path:   "foo.bar.value3",
			expect: "abc8080foo10bar",
		},
		{
			name:   "test ${foo${bar}}",
			path:   "foo.bar.value4",
			expect: "}",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := defaultResolver(data)
			assert.NoError(t, err)
			rd := reader{
				values: data,
			}
			if v, ok := rd.Value(test.path); ok {
				var actual interface{}
				switch test.expect.(type) {
				case int:
					if actual, err = v.Int(); err == nil {
						assert.Equal(t, test.expect, int(actual.(int64)), "int value should be equal")
					}
				case string:
					if actual, err = v.String(); err == nil {
						assert.Equal(t, test.expect, actual, "string value should be equal")
					}
				case bool:
					if actual, err = v.Bool(); err == nil {
						assert.Equal(t, test.expect, actual, "bool value should be equal")
					}
				case float64:
					if actual, err = v.Float(); err == nil {
						assert.Equal(t, test.expect, actual, "float64 value should be equal")
					}
				default:
					actual = v.Load()
					if !reflect.DeepEqual(test.expect, actual) {
						t.Logf("expect: %#v, actural: %#v", test.expect, actual)
						t.Fail()
					}
				}
				if err != nil {
					t.Error(err)
				}
			} else {
				t.Error("value path not found")
			}
		})
	}
}
