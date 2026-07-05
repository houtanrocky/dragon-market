package order

import "errors"

var ErrOrderNotFound = errors.New("order not found")
var ErrInvalidOrder = errors.New("order invalid")
var ErrOwnerNotFound = errors.New("owner not found")
