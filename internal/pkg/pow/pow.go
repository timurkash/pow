package pow

import (
	"crypto/sha256"
	"fmt"
	"log"
)

type HashCashData struct {
	Version    int
	ZerosCount int
	Date       int64
	Resource   string
	Rand       string
	Counter    int
	hash       []byte
}

func (h *HashCashData) Stringify() []byte {
	return []byte(fmt.Sprintf("%d:%d:%d:%s::%s:%d", h.Version, h.ZerosCount, h.Date, h.Resource, h.Rand, h.Counter))
}

func sha256Hash(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func (h *HashCashData) CalcHash() {
	h.hash = sha256Hash(h.Stringify())
}

func (h *HashCashData) PrintHash() {
	if h.hash == nil {
		log.Println("hash not calculated")
		return
	}
	fmt.Printf("%x\n", h.hash)
}

func (h *HashCashData) IsHashCorrect() bool {
	if h.hash == nil {
		return false
	}
	if h.ZerosCount > len(h.hash) {
		return false
	}
	for _, ch := range h.hash[:h.ZerosCount] {
		if ch != 0 {
			return false
		}
	}
	return true
}

func (h *HashCashData) ComputeHashCash(maxIterations int) (HashCashData, error) {
	for h.Counter <= maxIterations || maxIterations <= 0 {
		h.CalcHash()
		if h.IsHashCorrect() {
			return *h, nil
		}
		// if hash don't have needed count of leading zeros, we are increasing counter and try next hash
		h.Counter++
	}
	return *h, fmt.Errorf("max iterations exceeded")
}
