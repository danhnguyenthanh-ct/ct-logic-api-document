package entity

type AdListing struct {
	AdId       int64  `json:"ad_id"`
	ListId     int64  `json:"list_id"`
	CategoryId int64  `json:"category"`
	Body       string `json:"body"`
	Subject    string `json:"subject"`
}

type AdListingParam struct {
	Address  string `json:"address"`
	Area     string `json:"area"`
	Ward     string `json:"ward"`
	Region   string `json:"region"`
	Price    string `json:"price"`
	CarBrand string `json:"car_brand"`
	CarModel string `json:"car_model"`
	GearBox  string `json:"gear_box"`
	MfDate   string `json:"mf_date"`
}

type Ad struct {
	Name       string `json:"name"`
	AdID       int64  `json:"ad_id"`
	AccountID  int64  `json:"account_id"`
	CategoryID int64  `json:"category"`
	ListID     int64  `json:"list_id"`
	ListTime   string `json:"list_time"`
	CompanyAd  bool   `json:"company_ad"`
	Status     string `json:"status"`
	Type       string `json:"type"`
	Image      string `json:"image"`
	Pass       string `json:"salted_passwd"`
	Price      int64  `json:"price"`
	Body       string `json:"body"`
}
