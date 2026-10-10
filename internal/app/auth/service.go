package auth

import (
	"context"
	"unicode/utf8"

	"github.com/miebyte/authkit"
)

// Service 只在注册用例中允许 authkit 新建账号。
type Service struct {
	*authkit.Service
	store authkit.Store
}

func NewService(
	store authkit.Store,
	sender authkit.CodeSender,
	config authkit.Config,
) (*Service, error) {
	service, err := authkit.NewService(store, sender, nil, nil, config)
	if err != nil {
		return nil, err
	}
	return &Service{Service: service, store: store}, nil
}

// Register 原子地验证邮箱、创建账号、设置密码并签发有效会话。
func (s *Service) Register(
	ctx context.Context,
	input RegisterRequest,
	ip string,
) (*authkit.LoginResult, error) {
	length := utf8.RuneCountInString(input.Password)
	if !utf8.ValidString(input.Password) || length < 8 || length > 128 || len(ip) > 512 {
		return nil, authkit.ErrInvalidInput
	}
	var outcome authkit.Outcome
	err := s.store.WithTransaction(ctx, func(repos authkit.Repositories) error {
		policy := authkit.RegistrationPolicyFunc(
			func(context.Context, authkit.Registration) error { return nil },
		)
		tx := s.InTransaction(repos, policy)
		var err error
		outcome, err = tx.LoginEmail(
			ctx,
			authkit.EmailLoginInput{Email: input.Email, Code: input.Code},
		)
		// 验证拒绝必须提交，保留验证码的错误尝试次数。
		if err != nil || outcome.Rejected != nil {
			return err
		}
		if !outcome.Login.Created {
			return authkit.ErrConflict
		}
		if err := tx.SetPassword(ctx, authkit.SetPasswordInput{
			AccountID:  outcome.Login.Account.ID,
			Identifier: outcome.Login.Account.Email,
			Password:   input.Password,
		}); err != nil {
			return err
		}
		// 设置密码会撤销已有会话，之后才签发交付给客户端的会话。
		outcome, err = tx.LoginPassword(ctx, authkit.PasswordLoginInput{
			Identifier: outcome.Login.Account.Email, Password: input.Password, IP: ip,
		})
		if outcome.Rejected != nil {
			return outcome.Rejected
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	if outcome.Rejected != nil {
		return nil, outcome.Rejected
	}
	outcome.Login.Created = true
	return outcome.Login, nil
}
