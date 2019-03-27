package logext

import (
	"context"

	"github.com/runner-mei/loong"
	"github.com/three-plus-three/modules/toolbox"
)

func ContextWithUser(ctx context.Context, u toolbox.User) context.Context {
	return loong.ContextWithUser(ctx, u)
}

func UserFromContext(ctx context.Context) toolbox.User {
	u := loong.UserFromContext(ctx)
	if u == nil {
		return nil
	}
	user, _ := u.(toolbox.User)
	return user
}

func UserIDFromContext(ctx context.Context) int64 {
	u := UserFromContext(ctx)
	if u == nil {
		return 0
	}
	return u.ID()
}

func UsernameFromContext(ctx context.Context) string {
	u := UserFromContext(ctx)
	if u == nil {
		return ""
	}
	return u.Name()
}

func UsernicknameFromContext(ctx context.Context) string {
	u := UserFromContext(ctx)
	if u == nil {
		return ""
	}
	return u.Nickname()
}
