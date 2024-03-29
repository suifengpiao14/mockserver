package mockserver

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/suifengpiao14/kvstruct"
	"github.com/tidwall/gjson"
)

//DiffJson 对比json 输出从original 转换为target 需要的json patch,方便查看json是否一致
func DiffJson(original []byte, target []byte) (patch []byte, err error) {
	if len(original) == 0 {
		return target, nil
	}
	if len(target) == 0 {
		return target, nil
	}
	patch, err = jsonpatch.CreateMergePatch(original, target)
	return
}

func RequestBody2Feature(r *http.Request) (feature Feature, err error) {
	if r.Body == nil {
		return nil, nil
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(b)) //重新填写
	kvs := kvstruct.JsonToKVS(string(b), "")
	feature = KVS2Feature(kvs)
	return feature, nil
}

func Request2Feature(r *http.Request) (feature Feature, err error) {
	kvs := kvstruct.KVS{}
	path := formatPath(r.URL.Path)
	kvs.AddReplace(
		kvstruct.KV{
			Key:   "host",
			Value: r.Host,
		},
		kvstruct.KV{
			Key:   "path",
			Value: path,
		},
		kvstruct.KV{
			Key:   "method",
			Value: r.Method,
		},
		kvstruct.KV{
			Key:   "query",
			Value: r.URL.RawQuery,
		},
	)
	for k, values := range r.Header {
		for i, value := range values {
			kvs.AddReplace(kvstruct.KV{
				Key:   fmt.Sprintf("header.%s.%d", k, i),
				Value: value,
			})
		}
	}

	if r.Body != nil {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(b)) //重新填写
		key := "body"
		if gjson.ValidBytes(b) {
			subKvs := kvstruct.JsonToKVS(string(b), key)
			kvs.AddReplace(subKvs...)
		} else {
			kvs.AddReplace(kvstruct.KV{
				Key:   key,
				Value: string(b),
			})

		}
	}

	feature = KVS2Feature(kvs)
	return feature, nil
}
