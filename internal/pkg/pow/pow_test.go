package pow

import (
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHash(t *testing.T) {
	{
		hash := sha256Hash([]byte("testdatalong 1231378612"))
		assert.Equal(t,
			"bf63f065d706e572452cb373f9c444f4feb30c04c65bb40b1dac6ae3529a7187",
			fmt.Sprintf("%x", hash),
		)
	}
	{
		hash := sha256Hash([]byte("super800"))
		assert.Equal(t,
			"3b3e9ce0f2efd2631e7bfb7f6cbe4e3a48bcf494ff2ce8f11ba00494c9070ad2",
			fmt.Sprintf("%x", hash),
		)
	}
}

func TestIsHashCorrect(t *testing.T) {
	hashCashData := &HashCashData{
		Version:    1,
		ZerosCount: 2,
		Date:       1658309851,
		Resource:   "hashCash",
		Rand:       "KAXvYYVlDs",
		Counter:    37314,
	}
	hash := hashCashData.GetHash()
	assert.Equal(t,
		"000083a7d79d37dc901ee43273dfa412c346156fe2d556536c73c08cebcfeaca",
		fmt.Sprintf("%x", hash),
	)
	assert.True(t, hashCashData.IsHashCorrect(hash))
	hashCashData.ZerosCount = 3
	assert.False(t, hashCashData.IsHashCorrect(hash))
}

func TestComputeHashCash(t *testing.T) {
	t.Parallel()

	t.Run("4 zeros", func(t *testing.T) {
		date := time.Date(2022, 3, 13, 2, 28, 0, 0, time.UTC)
		input := HashCashData{
			Version:    1,
			ZerosCount: 4,
			Date:       date.Unix(),
			Resource:   "some_useful_data",
			Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", 123459))),
			Counter:    0,
		}
		result, err := input.ComputeHashCash(-1)
		require.NoError(t, err)
		assert.Equal(t, 26394, result.Counter)
	})
	t.Run("5 zeros", func(t *testing.T) {
		date := time.Date(2022, 3, 13, 2, 30, 0, 0, time.UTC)
		input := HashCashData{
			Version:    1,
			ZerosCount: 5,
			Date:       date.Unix(),
			Resource:   "some_useful_data",
			Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", 123460))),
			Counter:    0,
		}
		result, err := input.ComputeHashCash(-1)
		require.NoError(t, err)
		assert.Equal(t, 36258, result.Counter)
	})
	t.Run("Impossible challenge", func(t *testing.T) {
		date := time.Date(2022, 3, 13, 2, 30, 0, 0, time.UTC)
		input := HashCashData{
			Version:    1,
			ZerosCount: 10,
			Date:       date.Unix(),
			Resource:   "some_useful_data",
			Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", 123460))),
			Counter:    0,
		}
		result, err := input.ComputeHashCash(10)
		require.Error(t, err)
		assert.Equal(t, 11, result.Counter)
		assert.Equal(t, "max iterations exceeded", err.Error())
	})
}
