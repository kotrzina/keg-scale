package shops

type StockType int

const (
	StockTypeAvailable = iota
	StockTypeUnknown
)

type Beer struct {
	Title string // name of the item
	Price int    // price in czk
	Stock StockType
}
