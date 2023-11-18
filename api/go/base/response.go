package base

type DataField map[string]interface{} // struct to hold key-value pairs for json response

type Response struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Data    DataField `json:"data,omitempty"`
}
