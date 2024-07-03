package task

type Timezone string
type Status string

const (
	UTC Timezone = "utc"

	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
	StatusExpired  Status = "expired" // when account is expired
)

type Task struct {
	ID          string   `json:"id"`
	GroupID     string   `json:"group_id"`
	Name        string   `json:"name"`
	Expression  string   `json:"expression"`
	Timezone    Timezone `json:"timezone"`
	Timeout     int      `json:"timeout"`
	Instances   int      `json:"instances"`
	URL         string   `json:"url"`
	HTTPMethod  string   `json:"http_method"`
	HTTPHeaders string   `json:"http_headers"`
	Payload     string   `json:"payload"`
	Failures    int      `json:"failures"`
	Notify      bool     `json:"notify"`
	NotifyAfter int      `json:"notify_after"`
	Points      int      `json:"points"`
	Status      Status   `json:"status"`
}

type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
