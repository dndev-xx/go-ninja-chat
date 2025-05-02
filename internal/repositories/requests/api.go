package requests

import (
	"context"

	"github.com/dndev-xx/go-ninja-chat/internal/store/request"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func (r *Repo) CreateIfNotExists(ctx context.Context, requestID types.RequestID) (bool, error) {
	exitst, err := r.db.Request(ctx).
		Query().
		Where(request.ID(requestID)).
		Exist(ctx)
	if err != nil {
		return false, err
	}
	if exitst {
		return false, nil
	}
	_, err = r.db.Request(ctx).
		Create().
		SetID(requestID).
		Save(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}
