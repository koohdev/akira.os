package domain

type BatchJob struct {
	Ticker   string `json:"ticker"`
	Theme    string `json:"theme"`
	Platform string `json:"platform"`
	Twitter  string `json:"twitter"`
	Telegram string `json:"telegram"`
	Discord  string `json:"discord"`
	Dex      string `json:"dex"`
	Assets   string `json:"assets"`
	Context  string `json:"context"`
}