package services

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"

	"github.com/miebyte/goutils/utils"
	"github.com/superwhys/rent-manager-server/config"
	"github.com/superwhys/rent-manager-server/internal/domain/account"
	"github.com/superwhys/rent-manager-server/internal/errcode"
	"github.com/superwhys/rent-manager-server/internal/infra/repo"
	"github.com/superwhys/rent-manager-server/internal/jwtutils"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
	"github.com/superwhys/rent-manager-server/internal/models/response"
)

type AuthService struct {
	jwtConfig    *config.JwtConfig
	wechatConfig *config.WechatConfig
	accountRepo  *repo.AccountRepository
}

func NewAuthService(
	jwtConfig *config.JwtConfig,
	wechatConfig *config.WechatConfig,
	accountRepo *repo.AccountRepository,
) *AuthService {
	return &AuthService{
		jwtConfig:    jwtConfig,
		wechatConfig: wechatConfig,
		accountRepo:  accountRepo,
	}
}

func (a *AuthService) WechatLogin(ctx context.Context, req *requests.WechatLoginReq) (*response.TokenResponse, error) {
	wechatURL, err := url.Parse(a.wechatConfig.Jscode2SessionApi)
	if err != nil {
		return nil, fmt.Errorf("parse jscode2SessionApiUrl: %w", err)
	}

	q := wechatURL.Query()
	q.Add("appid", a.wechatConfig.AppID)
	q.Add("secret", a.wechatConfig.SecretID)
	q.Add("grant_type", "authorization_code")
	q.Add("js_code", req.Code)

	wechatURL.RawQuery = q.Encode()

	resp, err := http.Get(wechatURL.String())
	if err != nil {
		return nil, fmt.Errorf("getSessionKeyResponse: %w", err)
	}
	defer resp.Body.Close()

	type wechatSessionKeyResponse struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		UnionID    string `json:"unionid,omitempty"`
		ErrCode    int    `json:"errcode,omitempty"`
		ErrMsg     string `json:"errmsg,omitempty"`
	}

	sessionResp := &wechatSessionKeyResponse{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(sessionResp); err != nil {
		return nil, fmt.Errorf("decodeSessionResp: %w", err)
	}

	if sessionResp.ErrCode != 0 {
		return nil, fmt.Errorf("code=%s, errcode=%d, errmsg=%s", req.Code, sessionResp.ErrCode, sessionResp.ErrMsg)
	}

	acc, err := a.wechatLogin(ctx, sessionResp.OpenID)
	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, _, err := jwtutils.GenerateTokenPair(
		[]byte(a.jwtConfig.JwtSecret),
		a.jwtConfig.JwtTimeout,
		a.jwtConfig.JwtRefreshTimeout,
		acc.ID,
	)
	if err != nil {
		return nil, err
	}

	return &response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (*response.TokenResponse, error) {
	claims, err := jwtutils.ParseToken(refreshToken, []byte(a.jwtConfig.JwtSecret))
	if err != nil {
		if errors.Is(err, jwtutils.ErrTokenExpired) {
			return nil, errcode.ErrTokenExpired
		}
		return nil, errcode.ErrInvalidToken
	}
	if claims.TokenType != jwtutils.TokenTypeRefresh || claims.Subject == "" {
		return nil, errcode.ErrUnauthorized
	}

	// TODO: check token subject sessionID

	accessToken, err := jwtutils.GenerateAccessToken(
		[]byte(a.jwtConfig.JwtSecret),
		a.jwtConfig.JwtTimeout,
		claims.UserID,
		claims.Subject,
	)
	if err != nil {
		return nil, err
	}

	return &response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *AuthService) wechatLogin(ctx context.Context, openID string) (*account.Account, error) {
	if openID == "" {
		return nil, errcode.ErrBadRequest
	}

	acc, err := a.accountRepo.GetByWechatOpenID(ctx, openID)
	if err != nil {
		return nil, err
	}
	if acc != nil {
		return acc, nil
	}

	suffix, err := generateRandomAlnum(6)
	if err != nil {
		return nil, err
	}
	user := &account.Account{
		WechatOpenID: openID,
		Name:         "老板" + suffix,
	}
	return a.accountRepo.Create(ctx, user)
}

func generateRandomAlnum(length int) (string, error) {
	const charset = utils.Numeral + utils.UpperLetters
	if length <= 0 {
		return "", errcode.ErrBadRequest
	}

	buf := make([]byte, length)
	for i := range length {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		buf[i] = charset[n.Int64()]
	}

	return string(buf), nil
}
