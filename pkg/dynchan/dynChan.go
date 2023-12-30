package dynchan

func New[T any]() (chan<- T, <-chan T) {
	in := make(chan T)
	out := make(chan T)
	go dynChanLoop(in, out)
	return in, out
}

func NewBuffered[T any](bufferSize uint) (chan<- T, <-chan T) {
	in := make(chan T, bufferSize)
	out := make(chan T, bufferSize)
	go dynChanLoop(in, out)
	return in, out
}

// Credits: Inspired from the following sites and posts, although modified, fixed and completed.
// - https://medium.com/capital-one-tech/building-an-unbounded-channel-in-go-789e175cd2cd
// - https://stackoverflow.com/questions/41906146/why-go-channels-limit-the-buffer-size
func dynChanLoop[T any](in chan T, out chan T) {
	var nextValue T
	var inQueue []T

	defer close(out)

	for len(inQueue) > 0 || in != nil {
		if len(inQueue) == 0 {
			// Wait for a new value (nothing to serve, no need to sync between in and out channels with `select`).
			valueIn, ok := <-in
			if !ok {
				return // input channel is closed & there is no value left to serve: done
			}
			inQueue = append(inQueue, valueIn)
		}

		// At this point, there is at least 1 value in the queue.
		// Select on either new values coming in or requests to send next value in queue.
		nextValue = inQueue[0]
		select {
		case valueIn, ok := <-in:
			if !ok {
				in = nil // causes `in` to no longer be read by `select` statements
			} else {
				inQueue = append(inQueue, valueIn)
			}
		case out <- nextValue:
			inQueue = inQueue[1:] // pop first element out
		}
	}
}
