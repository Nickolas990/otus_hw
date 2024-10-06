package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	//nolint:depguard
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")

		require.False(t, ok)

		_, ok = c.Get("bbb")

		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)

		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)

		require.False(t, wasInCache)

		val, ok := c.Get("aaa")

		require.True(t, ok)

		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")

		require.True(t, ok)

		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)

		require.True(t, wasInCache)

		val, ok = c.Get("aaa")

		require.True(t, ok)

		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")

		require.False(t, ok)

		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		c := NewCache(3)

		// Добавляем 3 элемента
		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		wasInCache = c.Set("ccc", 300)
		require.False(t, wasInCache)

		// Проверяем, что все три элемента в кэше
		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		val, ok = c.Get("ccc")
		require.True(t, ok)
		require.Equal(t, 300, val)

		// Добавляем 4-й элемент, который должен вытеснить самый старый (aaa)
		wasInCache = c.Set("ddd", 400)
		require.False(t, wasInCache)

		// Проверяем, что "aaa" был вытеснен, таким образом проверяем, оба кейса: на размер кэша и старейший элемент
		val, ok = c.Get("aaa")
		require.False(t, ok)
		require.Nil(t, val)

		// Проверяем, что остальные элементы остались в кэше
		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		val, ok = c.Get("ccc")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ddd")
		require.True(t, ok)
		require.Equal(t, 400, val)
	})
}

func TestCacheUpdate(t *testing.T) {
	c := NewCache(3)

	// Добавляем элементы
	c.Set("key1", "val1")
	c.Set("key2", "val2")
	c.Set("key3", "val3")

	wasInCache := c.Set("key2", "new_val2")
	require.True(t, wasInCache)

	val, ok := c.Get("key2")
	require.True(t, ok)
	require.Equal(t, "new_val2", val)

	val, ok = c.Get("key1")
	require.True(t, ok)
	require.Equal(t, "val1", val)

	val, ok = c.Get("key3")
	require.True(t, ok)
	require.Equal(t, "val3", val)
}

func TestCacheClear(t *testing.T) {
	c := NewCache(3)

	c.Set("key1", "val1")
	c.Set("key2", "val2")
	c.Set("key3", "val3")

	c.Clear()

	val, ok := c.Get("key1")
	require.False(t, ok)
	require.Nil(t, val)

	val, ok = c.Get("key2")
	require.False(t, ok)
	require.Nil(t, val)

	val, ok = c.Get("key3")
	require.False(t, ok)
	require.Nil(t, val)
}

func TestEmptyCache(t *testing.T) {
	c := NewCache(3)

	// Проверяем, что запросы к пустому кэшу возвращают nil и false
	val, ok := c.Get("key1")
	require.False(t, ok)
	require.Nil(t, val)

	// Проверяем, что добавление и удаление работает корректно
	wasInCache := c.Set("key1", "val1")
	require.False(t, wasInCache)

	val, ok = c.Get("key1")
	require.True(t, ok)
	require.Equal(t, "val1", val)
}

func TestCacheMultithreading(_ *testing.T) {
	c := NewCache(10)

	wg := &sync.WaitGroup{}

	wg.Add(2)

	go func() {
		defer wg.Done()

		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()

		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}
