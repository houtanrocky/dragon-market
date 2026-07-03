package postgres

//
//import (
//	"context"
//	"database/sql"
//	"market-dragon/internal/auction"
//)
//
//type AuctionRepository struct {
//	db *sql.DB
//}
//
//func NewAuctionRepository(db *sql.DB) *AuctionRepository {
//	return &AuctionRepository{db: db}
//}
//
//func (r *AuctionRepository) Create(ctx context.Context, a *auction.Auction) error {
//	panic("implement me")
//}
//
//func (r *AuctionRepository) GetByID(ctx context.Context, id string) (*auction.Auction, error) {
//	panic("implement me")
//}
//
//func (r *AuctionRepository) GetActiveByItemID(ctx context.Context, itemID string) (*auction.Auction, error) {
//	panic("implement me")
//}
//
//func (r *AuctionRepository) Update(ctx context.Context, a *auction.Auction) error {
//	panic("implement me")
//}
//
//func (r *AuctionRepository) PlaceBid(ctx context.Context, b *auction.Bid) error {
//	panic("implement me")
//}
//
//func (r *AuctionRepository) GetTopBid(ctx context.Context, auctionID string) (*auction.Bid, error) {
//	panic("implement me")
//}
//
//func (r *AuctionRepository) GetBidsByAuction(ctx context.Context, auctionID string) ([]*auction.Bid, error) {
//	panic("implement me")
//}
//
//func (r *AuctionRepository) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
//	panic("implement me")
//}
