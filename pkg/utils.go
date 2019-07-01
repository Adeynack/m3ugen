package pkg

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
