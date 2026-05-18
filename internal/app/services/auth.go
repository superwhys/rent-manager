package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/superwhys/rent-manager-server/config"
	"github.com/superwhys/rent-manager-server/internal/domain/account"
	"github.com/superwhys/rent-manager-server/internal/errcode"
	"github.com/superwhys/rent-manager-server/internal/jwtutils"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
	"github.com/superwhys/rent-manager-server/internal/models/response"
)

type AuthService struct {
	jwtConfig    *config.JwtConfig
	wechatConfig *config.WechatConfig
	accountRepo  account.Repository
	httpClient   *http.Client
}

func NewAuthService(jwtConfig *config.JwtConfig, wechatConfig *config.WechatConfig, accountRepo account.Repository) *AuthService {
	return &AuthService{
		jwtConfig:    jwtConfig,
		wechatConfig: wechatConfig,
		accountRepo:  accountRepo,
		httpClient:   http.DefaultClient,
	}
}

func (s *AuthService) WechatLogin(ctx context.Context, req *requests.WechatLoginReq) (*response.TokenResponse, error) {
	if req.Code == "" {
		return nil, errcode.ErrBadRequest.WithMessage("微信登录 code 不能为空")
	}

	openID, err := s.fetchWechatOpenID(ctx, req.Code)
	if err != nil {
		return nil, err
	}

	accountInfo, err := s.accountRepo.GetByWechatOpenID(ctx, openID)
	if err != nil {
		if !errors.Is(err, account.ErrAccountNotFound) {
			return nil, err
		}
		accountInfo, err = account.NewWechatAccount(openID)
		if err != nil {
			return nil, err
		}
		accountInfo, err = s.accountRepo.Create(ctx, accountInfo)
		if err != nil {
			return nil, err
		}
	}

	accessToken, refreshToken, _, err := jwtutils.GenerateTokenPair(
		[]byte(s.jwtConfig.JwtSecret),
		s.jwtConfig.JwtTimeout,
		s.jwtConfig.JwtRefreshTimeout,
		accountInfo.ID,
	)
	if err != nil {
		return nil, err
	}

	return &response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *requests.RefreshTokenReq) (*response.TokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, errcode.ErrBadRequest.WithMessage("刷新令牌不能为空")
	}

	claims, err := jwtutils.ParseToken(req.RefreshToken, []byte(s.jwtConfig.JwtSecret))
	if err != nil {
		return nil, errcode.ErrInvalidToken
	}
	if claims.TokenType != jwtutils.TokenTypeRefresh {
		return nil, errcode.ErrInvalidToken
	}

	accessToken, err := jwtutils.GenerateAccessToken(
		[]byte(s.jwtConfig.JwtSecret),
		s.jwtConfig.JwtTimeout,
		claims.UserID,
		claims.Subject,
	)
	if err != nil {
		return nil, err
	}

	return &response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
	}, nil
}

type wechatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func (s *AuthService) fetchWechatOpenID(ctx context.Context, code string) (string, error) {
	apiURL, err := url.Parse(s.wechatConfig.Jscode2SessionApi)
	if err != nil {
		return "", err
	}

	query := apiURL.Query()
	query.Set("appid", s.wechatConfig.AppID)
	query.Set("secret", s.wechatConfig.SecretID)
	query.Set("js_code", code)
	query.Set("grant_type", "authorization_code")
	apiURL.RawQuery = query.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return "", err
	}

	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("微信登录接口返回异常状态码 %d", httpResp.StatusCode)
	}

	var sessionResp wechatSessionResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&sessionResp); err != nil {
		return "", err
	}
	if sessionResp.ErrCode != 0 {
		return "", errcode.ErrBadRequest.WithMessage(sessionResp.ErrMsg)
	}
	if sessionResp.OpenID == "" {
		return "", errcode.ErrBadRequest.WithMessage("微信 openid 为空")
	}

	return sessionResp.OpenID, nil
}
