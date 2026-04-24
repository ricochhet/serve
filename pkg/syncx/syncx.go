package syncx

import "github.com/sasha-s/go-deadlock"

type Safe[T any] struct {
	Mutex deadlock.Mutex
	t     *T
}

// NewSafe creates an empty Safe.
func NewSafe[T any]() *Safe[T] {
	return &Safe[T]{}
}

// GetLocked returns the target.
func (s *Safe[T]) GetLocked() *T {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	return s.t
}

// Get is the same as GetLocked() but does not lock the mutex.
func (s *Safe[T]) Get() *T {
	return s.t
}

// SetLocked sets the target.
func (s *Safe[T]) SetLocked(t *T) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.t = t
}

// Set is the same as SetLocked() but does not lock the mutex.
func (s *Safe[T]) Set(t *T) {
	s.t = t
}

// CopyFrom sets all target to the target.
func (s *Safe[T]) CopyFrom(target *Safe[T]) {
	s.SetLocked(target.GetLocked())
}

// GetMap returns the value for key from a Safe map.
func GetMap[K comparable, V any](s *Safe[map[K]V], key K) (V, bool) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	m := s.Get()
	if m == nil {
		var zero V
		return zero, false
	}

	v, ok := (*m)[key]

	return v, ok
}

// SetMap sets the value for key in a Safe map.
func SetMap[K comparable, V any](s *Safe[map[K]V], key K, value V) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	m := s.Get()
	if m == nil {
		s.Set(&map[K]V{key: value})
		return
	}

	(*m)[key] = value
}
