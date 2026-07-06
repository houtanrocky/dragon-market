package order

import "errors"

var ErrOrderNotFound = errors.New("order not found")
var ErrInvalidOrder = errors.New("order invalid")
var ErrOwnerNotFound = errors.New("owner not found")
var ErrOrderAlreadySold = errors.New("order is already sold")
var ErrOrderNotListed = errors.New("order is not already listed for sale")
var ErrOrderItemNotOwnedBySeller = errors.New("order item is not owned by seller")
var ErrOrderItemNotAvailable = errors.New("order item is not available")
var ErrCantListOrderListedInAuction = errors.New("cannot make limit order for item that's listed in auction")
var ErrCancelOrderListedByAnother = errors.New("cannot cancel order that's listed by another user")
var ErrCancelOrderNotListed = errors.New("cannot cancel order that's not listed")
var ErrCannotBuyOwnOrder = errors.New("cannot buy own order")
