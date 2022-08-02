package eth

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type BeaconBlockRequest struct {
	BlockID BlockIdentifier `json:"block_id"`
}

func (r *BeaconBlockRequest) Validate() error {
	if r.BlockID.Type() == Invalid {
		return errors.New("invalid block ID")
	}

	return nil
}

func NewValidatedBeaconBlockRequestFromRequest(r *http.Request, p httprouter.Params) (*BeaconBlockRequest, error) {
	id, err := NewBlockIdentifier(p.ByName("block_id"))
	if err != nil {
		return nil, err
	}

	req := &BeaconBlockRequest{
		BlockID: id,
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	return req, nil
}
