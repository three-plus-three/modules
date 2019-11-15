package userext

import (
	"context"

	"github.com/runner-mei/loong"
	"github.com/three-plus-three/modules/users"
)

func ContextWithUser(ctx context.Context, u users.ReadCurrentUserFunc) context.Context {
	return loong.ContextWithUser(ctx, u)
}

func UserFromContext(ctx context.Context) users.ReadCurrentUserFunc {
	o := loong.UserFromContext(ctx)
	if o == nil {
		return nil
	}
	f, _ := o.(users.ReadCurrentUserFunc)
	return f
}
