package utils

type ErrorResponse struct {
    Error string `json:"error"`
}

type OKResponse struct {
    Data interface{} `json:"data"`
}
