package endpoints

type schemaResponse struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
}
