package m3ugen

import "math/rand"

// FirstErr returns the first errors in a list.
func FirstErr(first, second error, others ...error) error {
	if first != nil {
		return first
	}
	if second != nil {
		return second
	}
	for _, e := range others {
		if e != nil {
			return e
		}
	}
	return nil
}

// ShuffleSlice takes a slice of any type and randomize the order of its content.
func ShuffleSlice[T any](a []T) {
	rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
}
