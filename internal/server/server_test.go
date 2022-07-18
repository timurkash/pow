package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timurkash/pow/internal/pkg/cache"
	"github.com/timurkash/pow/internal/pkg/config"
	"github.com/timurkash/pow/internal/pkg/pow"
	"github.com/timurkash/pow/internal/pkg/protocol"
	"strings"
	"testing"
	"time"
)

// MockClock - mock for Clock interface (to work with predefined Now)
type MockClock struct {
	NowFunc func() time.Time
}

func (m *MockClock) Now() time.Time {
	if m.NowFunc != nil {
		return m.NowFunc()
	}
	return time.Now()
}

func TestProcessRequest(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "config", &config.Config{HashCashZerosCount: 3, HashCashDuration: 300})
	mockClock := MockClock{}
	ctx = context.WithValue(ctx, "clock", &mockClock)
	cacheInst := cache.InitInMemoryCache(&mockClock)
	ctx = context.WithValue(ctx, "cache", cacheInst)

	const randKey = 123460

	t.Run("Quit request", func(t *testing.T) {
		input := fmt.Sprintf("%d|", protocol.Quit)
		msg, err := ProcessRequest(ctx, input, "client1")
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, ErrQuit, err)
	})

	t.Run("Invalid request", func(t *testing.T) {
		input := "||"
		msg, err := ProcessRequest(ctx, input, "client1")
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "message doesn't match protocol", err.Error())
	})

	t.Run("Unknown header", func(t *testing.T) {
		input := "111|"
		msg, err := ProcessRequest(ctx, input, "client1")
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "unknown header", err.Error())
	})

	t.Run("Request challenge", func(t *testing.T) {
		input := fmt.Sprintf("%d|", protocol.RequestChallenge)
		msg, err := ProcessRequest(ctx, input, "client1")
		require.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, protocol.ResponseChallenge, msg.Header)
		//unmarshal msg to check fields
		var hashCash pow.HashCashData
		err = json.Unmarshal([]byte(msg.Payload), &hashCash)
		require.NoError(t, err)
		assert.Equal(t, 3, hashCash.ZerosCount)
		assert.Equal(t, "client1", hashCash.Resource)
		assert.NotEmpty(t, hashCash.Rand)
	})

	t.Run("Request resource without solution", func(t *testing.T) {
		input := fmt.Sprintf("%d|", protocol.RequestResource)
		msg, err := ProcessRequest(ctx, input, "client1")
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.True(t, strings.Contains(err.Error(), "err unmarshal hashcash"))
	})

	t.Run("Request resource with wrong resource", func(t *testing.T) {
		hashCash := pow.HashCashData{
			Version:    1,
			ZerosCount: 4,
			Date:       time.Now().Unix(),
			Resource:   "client2",
			Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", randKey))),
			Counter:    100,
		}
		marshaled, err := json.Marshal(hashCash)
		require.NoError(t, err)
		input := fmt.Sprintf("%d|%s", protocol.RequestResource, string(marshaled))
		msg, err := ProcessRequest(ctx, input, "client1")
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "invalid hashCash resource", err.Error())
	})

	t.Run("Request resource with invalid solution and 0 counter", func(t *testing.T) {
		cacheInst.Add(randKey, 100)

		hashCash := pow.HashCashData{
			Version:    1,
			ZerosCount: 10,
			Date:       time.Now().Unix(),
			Resource:   "client1",
			Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", randKey))),
			Counter:    0,
		}
		marshaled, err := json.Marshal(hashCash)
		require.NoError(t, err)
		input := fmt.Sprintf("%d|%s", protocol.RequestResource, string(marshaled))
		msg, err := ProcessRequest(ctx, input, "client1")
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "invalid hashCash", err.Error())
	})

	t.Run("Request resource with expired solution", func(t *testing.T) {
		mockClock.NowFunc = func() time.Time {
			return time.Date(2022, 3, 13, 2, 30, 0, 0, time.UTC)
		}
		cacheInst.Add(randKey, 100)

		mockClock.NowFunc = func() time.Time {
			return time.Date(2022, 3, 13, 2, 40, 0, 0, time.UTC)
		}
		hashCash := pow.HashCashData{
			Version:    1,
			ZerosCount: 10,
			Date:       time.Date(2022, 3, 13, 2, 30, 0, 0, time.UTC).Unix(),
			Resource:   "client1",
			Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", randKey))),
			Counter:    100,
		}
		marshaled, err := json.Marshal(hashCash)
		require.NoError(t, err)
		input := fmt.Sprintf("%d|%s", protocol.RequestResource, string(marshaled))
		msg, err := ProcessRequest(ctx, input, "client1")
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "challenge expired or not sent", err.Error())
	})

	t.Run("Request resource with correct solution", func(t *testing.T) {
		mockClock.NowFunc = func() time.Time {
			return time.Date(2022, 3, 13, 2, 30, 0, 0, time.UTC)
		}
		err := cacheInst.Add(randKey, 200)
		assert.NoError(t, err)

		mockClock.NowFunc = func() time.Time {
			return time.Date(2022, 3, 13, 2, 32, 0, 0, time.UTC)
		}
		date := time.Date(2022, 3, 13, 2, 30, 0, 0, time.UTC)
		hashCash := pow.HashCashData{
			Version:    1,
			ZerosCount: 3,
			Date:       date.Unix(),
			Resource:   "client1",
			Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", randKey))),
			Counter:    5001,
		}
		marshaled, err := json.Marshal(hashCash)
		require.NoError(t, err)
		input := fmt.Sprintf("%d|%s", protocol.RequestResource, string(marshaled))
		msg, err := ProcessRequest(ctx, input, "client1")
		require.NoError(t, err)
		assert.NotNil(t, msg)
		//assert.Contains(t, Quotes, msg.Payload)

		// check that rand key was deleted from cache
		exists, err := cacheInst.Get(randKey)
		assert.NoError(t, err)
		assert.False(t, exists)

		// request server second time
		msg, err = ProcessRequest(ctx, input, "client1")
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Equal(t, "challenge expired or not sent", err.Error())
	})
}
