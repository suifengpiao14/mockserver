package mockserver

import (
	jsonpatch "github.com/evanphx/json-patch/v5"
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
