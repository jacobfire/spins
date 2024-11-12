package server

type TokenResult struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwbGF5ZXIiLCJleHAiOjE3MzE0MjIyNDIsImlhdCI6MTczMTQxODY0MiwiaXNzIjoiYXV0aC1hcHAiLCJzdWIiOiJqb2huZG9lMUBleGFtcGxlLmNvbSJ9.pzwLR3DVS40YF4FheURIUDRLk0dyQvLg4-cUOykanqA"`
}
type JSONResult struct {
	Response string `json:"message"`
}
