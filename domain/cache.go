package domain

import (
	"image"
	"sync"
)

var C Cache = &memoryCache{m: make(map[string][2]string)}

type Cache interface {
	Get(url string) (string, string, bool, error)
	Set(url, text, image string) error
	GetImage(url string) *image.Image
	SetImage(url string, i *image.Image)
}

type memoryCache struct {
	mu  sync.RWMutex
	m   map[string][2]string
	img map[string]*image.Image
}

func (m *memoryCache) Get(url string) (string, string, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.m[url]
	return v[0], v[1], ok, nil
}

func (m *memoryCache) Set(url, text, image string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[url] = [2]string{text, image}
	return nil
}

func (m *memoryCache) GetImage(url string) *image.Image {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cachedImage, _ := m.img[url]
	return cachedImage
}

func (m *memoryCache) SetImage(url string, i *image.Image) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.img[url] = i
}
