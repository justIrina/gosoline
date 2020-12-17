package db_repo

import "fmt"

type DuplicateEntryError struct {
	Err error
}

func (e *DuplicateEntryError) Error() string {
	return fmt.Sprintf("duplicate entry: %s", e.Err.Error())
}

func (e *DuplicateEntryError) Is(err error) bool {
	_, ok := err.(*DuplicateEntryError)

	return ok
}

func (e *DuplicateEntryError) As(target interface{}) bool {
	targetErr, ok := target.(*DuplicateEntryError)

	if ok {
		*targetErr = *e
	}

	return ok
}

func (e *DuplicateEntryError) Unwrap() error {
	return e.Err
}
