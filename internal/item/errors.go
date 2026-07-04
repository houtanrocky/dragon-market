package item

import "errors"

var ErrItemNotFound = errors.New("item not found")
var ErrInvalidItem = errors.New("item invalid")
var ErrOwnerNotFound = errors.New("owner not found")
