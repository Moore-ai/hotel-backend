package obfuscate

import (
	"github.com/speps/go-hashids/v2"
)

// NewKey 从密码创建 Hashids 配置（salt）
func NewKey(password string) *hashids.HashID {
	hd := hashids.NewData()
	hd.Salt = password
	hd.MinLength = 8
	h, err := hashids.NewWithData(hd)
	if err != nil {
		panic("obfuscate: failed to create hashids: " + err.Error())
	}
	return h
}

// Encode 将 uint ID 加密为短字符串
func Encode(h *hashids.HashID, id uint) (string, error) {
	return h.Encode([]int{int(id)})
}

// Decode 将加密字符串解密为 uint ID
func Decode(h *hashids.HashID, code string) (uint, error) {
	ids, err := h.DecodeWithError(code)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	return uint(ids[0]), nil
}
