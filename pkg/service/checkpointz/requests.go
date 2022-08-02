package checkpointz

type StatusRequest struct {
}

func (r *StatusRequest) Validate() error {
	return nil
}

func NewStatusRequest() *StatusRequest {
	return &StatusRequest{}
}
