package gondor

import (
	"errors"
	"time"
)

func WaitFor(timeout int, predicate func() (bool, error)) error {
	start := time.Now().Second()
	for {
		// Force a 1s sleep
		time.Sleep(1 * time.Second)

		// If a timeout is set, and that's been exceeded, shut it down
		if timeout >= 0 && time.Now().Second()-start >= timeout {
			return errors.New("A timeout occurred")
		}

		// Execute the function
		satisfied, err := predicate()
		if err != nil {
			return err
		}
		if satisfied {
			return nil
		}
	}
}
