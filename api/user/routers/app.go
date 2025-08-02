package routers

import (
	"context"

	"github.com/superwhys/rent-manager-server/model"
)

type App interface {
	Login(ctx context.Context, req *model.LoginRequest) (*model.User, error)
	Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error)
	Logout(ctx context.Context, req *model.LogoutRequest) error
	ChangePassword(ctx context.Context, req *model.ChangePasswordRequest) error
	GetUserInfo(ctx context.Context, req *model.GetUserInfoRequest) (*model.User, error)
	UpdateUserInfo(ctx context.Context, req *model.UpdateUserInfoRequest) error
	DeleteUser(ctx context.Context, req *model.DeleteUserRequest) error
}
