package handler

import "errors"

var (
	ErrAPIEmptyBatch = errors.New("nothing to process, empty batch")
)
