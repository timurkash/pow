package pow

import (
	"crypto/sha256"
	"fmt"
)

const zeroByte = 48

type HashCashData struct {
	Version    int
	ZerosCount int
	Date       int64
	Resource   string
	Rand       string
	Counter    int
}

func (h HashCashData) Stringify() string {
	return fmt.Sprintf("%d:%d:%d:%s::%s:%d", h.Version, h.ZerosCount, h.Date, h.Resource, h.Rand, h.Counter)
}

func sha256Hash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func IsHashCorrect(hash string, zerosCount int) bool {
	if zerosCount > len(hash) {
		return false
	}
	for _, ch := range hash[:zerosCount] {
		if ch != zeroByte {
			return false
		}
	}
	return true
}

func (h HashCashData) ComputeHashCash(maxIterations int) (HashCashData, error) {
	for h.Counter <= maxIterations || maxIterations <= 0 {
		header := h.Stringify()
		hash := sha256Hash(header)
		//fmt.Println(header, hash)
		if IsHashCorrect(hash, h.ZerosCount) {
			return h, nil
		}
		// if hash don't have needed count of leading zeros, we are increasing counter and try next hash
		h.Counter++
	}
	return h, fmt.Errorf("max iterations exceeded")
}
