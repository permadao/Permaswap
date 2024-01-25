package schema

type CollectStats struct {
	Count      float64 `json:"count"`
	NumOwners  int64   `json:"num_owners"`
	FloorPrice float64 `json:"floor_price"`
}

type Collect struct {
	Slug       string       `json:"slug"`
	Name       string       `json:"name"`
	ImageUrl   string       `json:"image_url"`
	CreateTime string       `json:"created_date"`
	Stats      CollectStats `json:"stats"`
}

type ArNFT struct {
	ContractAddr   string `json:"contractAddr"`
	TokenId        string `json:"tokenId"`
	Owner          string `json:"owner"`
	Name           string `json:"name"`
	CollectionName string `json:"collectionName"`
	ImageUrl       string `json:"imageUrl"`
	Timestamp      int64  `json:"timestamp"`
	DataUrl        string `json:"dataUrl"`
}

type ResNFT struct {
	ContractAddr      string  `json:"contractAddr"`
	TokenId           string  `json:"tokenId"`
	PermaLink         string  `json:"permaLink"`
	Collection        Collect `json:"collection"`
	ImageUrl          string  `json:"imageUrl"`
	Name              string  `json:"name"`
	NameDes           string  `json:"nameDes"`
	Price             string  `json:"price"`
	PriceSymbol       string  `json:"priceSymbol"`
	TopOffer          string  `json:"topOffer"`
	TopOfferSymbol    string  `json:"topOfferSymbol"`
	MinOffer          string  `json:"minOffer"`
	MinOfferSymbol    string  `json:"minOfferSymbol"`
	AuctionType       string  `json:"auctionType"`
	Owner             string  `json:"owner"`
	OwnerLink         string  `json:"ownerLink"`
	CollectionName    string  `json:"collectionName"`
	CollectionNameDes string  `json:"collectionNameDes"`
	Timestamp         int64   `json:"timestamp"`
}
