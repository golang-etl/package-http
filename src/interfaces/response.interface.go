package interfaces

type Response struct {
	IsFile     bool
	FilePath   string
	FileName   string
	StatusCode int
	Headers    map[string]string
	Body       interface{}
}

type ResponseBodyErrorStackItem struct {
	File     string `json:"file"`
	FuncName string `json:"funcName"`
}

type ResponseBodyError struct {
	Message   string                        `json:"message"`
	Stack     *[]ResponseBodyErrorStackItem `json:"stack,omitempty"`
	ErrorCode string                        `json:"errorCode"`
}

type ResponseBodySuccess = map[string]interface{}
