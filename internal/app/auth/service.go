package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"git.superwhys.top/rent-manager/rent-manager-server/config"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/domain/account"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"
)

// Service 提供登录和令牌刷新用例。
type Service struct {
	wechatConfig *config.WechatConfig
	accountRepo  account.Repository
	tokenManager TokenManager
	httpClient   *http.Client
}

// NewService 创建认证服务。
func NewService(wechatConfig *config.WechatConfig, accountRepo account.Repository, tokenManager TokenManager) *Service {
	return &Service{
		wechatConfig: wechatConfig,
		accountRepo:  accountRepo,
		tokenManager: tokenManager,
		httpClient:   http.DefaultClient,
	}
}

// WechatLogin 使用微信 code 登录并签发令牌。
func (s *Service) WechatLogin(ctx context.Context, req *WechatLoginReq) (*TokenResult, error) {
	if req.Code == "" {
		return nil, xerr.ErrBadRequest.WithMessage("微信登录 code 不能为空")
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

	tokenPair, err := s.tokenManager.GenerateTokenPair(accountInfo.ID)
	if err != nil {
		return nil, err
	}

	return &TokenResult{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RefreshToken 使用 refresh token 换取新的 access token。
func (s *Service) RefreshToken(ctx context.Context, req *RefreshTokenReq) (*TokenResult, error) {
	if req.RefreshToken == "" {
		return nil, xerr.ErrBadRequest.WithMessage("刷新令牌不能为空")
	}

	claims, err := s.tokenManager.ParseToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, xerr.ErrTokenExpired) {
			return nil, xerr.ErrLoginExpired
		}
		return nil, xerr.ErrInvalidToken
	}

	if claims.TokenType != TokenTypeRefresh {
		return nil, xerr.ErrInvalidToken
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(claims.UserID, claims.Subject)
	if err != nil {
		return nil, err
	}

	return &TokenResult{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
	}, nil
}

// wechatSessionResponse 表示微信 jscode2session 接口响应。
type wechatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// fetchWechatOpenID 通过微信 code 获取 openid。
func (s *Service) fetchWechatOpenID(ctx context.Context, code string) (string, error) {
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
		return "", xerr.ErrBadRequest.WithMessage(sessionResp.ErrMsg)
	}
	if sessionResp.OpenID == "" {
		return "", xerr.ErrBadRequest.WithMessage("微信 openid 为空")
	}

	return sessionResp.OpenID, nil
}
