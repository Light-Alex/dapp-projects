package types

import "github.com/shopspring/decimal"

type ItemInfo struct {
	CollectionAddress string `json:"collection_address"`
	TokenID           string `json:"token_id"`
}

type ItemPriceInfo struct {
	CollectionAddress string          `json:"collection_address"`
	TokenID           string          `json:"token_id"`
	Maker             string          `json:"maker"`
	Price             decimal.Decimal `json:"price"`
	OrderStatus       int             `json:"order_status"`
}

type ItemOwner struct {
	CollectionAddress string `json:"collection_address"`
	TokenID           string `json:"token_id"`
	Owner             string `json:"owner"`
}

type ItemImage struct {
	CollectionAddress string `json:"collection_address"`
	TokenID           string `json:"token_id"`
	ImageUri          string `json:"image_uri"`
}

// ItemDetailInfo 表示单个NFT项目的详细信息。
// 该结构体包含了项目的基本信息、所属集合信息、图片和视频信息、最近成交价格、地板价、挂单信息和出价信息等。
type ItemDetailInfo struct {
	// ChainID 表示NFT所在的区块链ID。
	ChainID int `json:"chain_id"`
	// Name 表示NFT项目的名称。
	Name string `json:"name"`
	// CollectionAddress 表示NFT所属集合的地址。
	CollectionAddress string `json:"collection_address"`
	// CollectionName 表示NFT所属集合的名称。
	CollectionName string `json:"collection_name"`
	// CollectionImageURI 表示NFT所属集合的图片URI。
	CollectionImageURI string `json:"collection_image_uri"`
	// TokenID 表示NFT的唯一标识符。
	TokenID string `json:"token_id"`
	// ImageURI 表示NFT的图片URI。
	ImageURI string `json:"image_uri"`
	// VideoType 表示NFT的视频类型。
	VideoType string `json:"video_type"`
	// VideoURI 表示NFT的视频URI。
	VideoURI string `json:"video_uri"`
	// LastSellPrice 表示NFT的最近成交价格。
	LastSellPrice decimal.Decimal `json:"last_sell_price"`
	// FloorPrice 表示NFT所属集合的地板价。
	FloorPrice decimal.Decimal `json:"floor_price"`
	// OwnerAddress 表示NFT的所有者地址。
	OwnerAddress string `json:"owner_address"`
	// MarketplaceID 表示NFT所在的交易市场ID。
	MarketplaceID int `json:"marketplace_id"`

	// ListOrderID 表示NFT的挂单订单ID。
	ListOrderID string `json:"list_order_id"`
	// ListTime 表示NFT的挂单时间。
	ListTime int64 `json:"list_time"`
	// ListPrice 表示NFT的挂单价格。
	ListPrice decimal.Decimal `json:"list_price"`
	// ListExpireTime 表示NFT挂单的过期时间。
	ListExpireTime int64 `json:"list_expire_time"`
	// ListSalt 表示NFT挂单的盐值。
	ListSalt int64 `json:"list_salt"`
	// ListMaker 表示NFT挂单的创建者地址。
	ListMaker string `json:"list_maker"`

	// BidOrderID 表示NFT的出价订单ID。
	BidOrderID string `json:"bid_order_id"`
	// BidTime 表示NFT的出价时间。
	BidTime int64 `json:"bid_time"`
	// BidExpireTime 表示NFT出价的过期时间。
	BidExpireTime int64 `json:"bid_expire_time"`
	// BidPrice 表示NFT的出价价格。
	BidPrice decimal.Decimal `json:"bid_price"`
	// BidSalt 表示NFT出价的盐值。
	BidSalt int64 `json:"bid_salt"`
	// BidMaker 表示NFT出价的创建者地址。
	BidMaker string `json:"bid_maker"`
	// BidType 表示NFT出价的类型。
	BidType int64 `json:"bid_type"`
	// BidSize 表示NFT出价的数量。
	BidSize int64 `json:"bid_size"`
	// BidUnfilled 表示NFT出价的未成交数量。
	BidUnfilled int64 `json:"bid_unfilled"`
}

type ItemDetailInfoResp struct {
	Result interface{} `json:"result"`
}

type ListingInfo struct {
	MarketplaceId int32           `json:"marketplace_id"`
	Price         decimal.Decimal `json:"price"`
}

type TraitPrice struct {
	CollectionAddress string          `json:"collection_address"`
	TokenID           string          `json:"token_id"`
	Trait             string          `json:"trait"`
	TraitValue        string          `json:"trait_value"`
	Price             decimal.Decimal `json:"price"`
}

type ItemTopTraitResp struct {
	Result interface{} `json:"result"`
}
