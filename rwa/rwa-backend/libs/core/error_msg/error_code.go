package error_msg

type ErrorCode struct {
	Code int
	Msg  string
}

const (
	ErrInvalidRequestParams = "ErrInvalidRequestParams"
	ErrInternalServerError  = "ErrInternalServerError"
	ErrNotFound             = "ErrNotFound"

	ErrFailedToGetStockList   = "ErrFailedToGetStockList"
	ErrFailedToGetStockDetail = "ErrFailedToGetStockDetail"
	ErrStockNotFound          = "ErrStockNotFound"

	ErrFailedToGetCurrentPrice   = "ErrFailedToGetCurrentPrice"
	ErrFailedToGetHistoricalData = "ErrFailedToGetHistoricalData"
	ErrFailedToGetMarketClock    = "ErrFailedToGetMarketClock"
	ErrFailedToGetLatestQuote    = "ErrFailedToGetLatestQuote"
	ErrInvalidTimestampFormat    = "ErrInvalidTimestampFormat"
	ErrFailedToGetSnapshot       = "ErrFailedToGetSnapshot"
	ErrFailedToGetAssets         = "ErrFailedToGetAssets"
	ErrSymbolRequired            = "ErrSymbolRequired"
	ErrFailedToGetAsset          = "ErrFailedToGetAsset"

	ErrFailedToGetOrders          = "ErrFailedToGetOrders"
	ErrFailedToGetOrderDetail     = "ErrFailedToGetOrderDetail"
	ErrOrderNotFound              = "ErrOrderNotFound"
	ErrFailedToGetOrderExecutions = "ErrFailedToGetOrderExecutions"
)

var errorCodeMap = map[string]ErrorCode{
	ErrInvalidRequestParams: {Code: 1000, Msg: "Invalid request parameters"},
	ErrInternalServerError:  {Code: 1001, Msg: "Internal Server Error"},
	ErrNotFound:             {Code: 1002, Msg: "Data not found"},

	ErrFailedToGetStockList:   {Code: 2000, Msg: "Failed to get stock list"},
	ErrFailedToGetStockDetail: {Code: 2001, Msg: "Failed to get stock detail"},
	ErrStockNotFound:          {Code: 2002, Msg: "Stock not found"},

	ErrFailedToGetCurrentPrice:   {Code: 3000, Msg: "failed to get current price"},
	ErrFailedToGetHistoricalData: {Code: 3001, Msg: "Failed to get historical data"},
	ErrFailedToGetMarketClock:    {Code: 3002, Msg: "Failed to get market clock"},
	ErrFailedToGetLatestQuote:    {Code: 3003, Msg: "Failed to get latest quote"},
	ErrInvalidTimestampFormat:    {Code: 3004, Msg: "Invalid timestamp format"},
	ErrFailedToGetSnapshot:       {Code: 3005, Msg: "failed to get snapshot"},
	ErrFailedToGetAssets:         {Code: 3006, Msg: "Failed to get assets"},
	ErrSymbolRequired:            {Code: 3007, Msg: "Symbol is required"},
	ErrFailedToGetAsset:          {Code: 3008, Msg: "Failed to get asset"},

	ErrFailedToGetOrders:          {Code: 4000, Msg: "Failed to get orders"},
	ErrFailedToGetOrderDetail:     {Code: 4001, Msg: "Failed to get order detail"},
	ErrOrderNotFound:              {Code: 4002, Msg: "Order not found"},
	ErrFailedToGetOrderExecutions: {Code: 4003, Msg: "Failed to get order executions"},
}

func GetErrorCode(key string) (ErrorCode, bool) {
	code, ok := errorCodeMap[key]
	return code, ok
}
