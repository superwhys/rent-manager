package middlewares

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"

	"git.superwhys.top/rent-manager/rent-manager-server/api/response"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"
	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/logging"
	"github.com/ulule/limiter/v3"

	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
)

const (
	RateLimitExcludeKey = "rate_limit_excluded"
)

func keyGenerator(excludedPaths []string) func(c *gin.Context) string {
	return func(c *gin.Context) string {
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		token := c.GetHeader("Authorization")

		for _, p := range excludedPaths {
			if strings.HasPrefix(path, p) {
				return RateLimitExcludeKey
			}
		}

		hash := md5.New()
		hash.Write([]byte(token))
		hash.Write([]byte(path))
		hash.Write([]byte(clientIP))
		key := fmt.Sprintf("%x", hash.Sum(nil))

		logging.Debugc(c.Request.Context(), "limit path: %s clientIp: %s key: %s", path, clientIP, key)
		return key
	}
}

func RateLimitMiddleware(
	instance *limiter.Limiter,
	excludedPaths []string,
) gin.HandlerFunc {
	if instance == nil {
		logging.PanicError(fmt.Errorf("limiter instance is nil"))
	}

	return ginlimiter.NewMiddleware(
		instance,
		ginlimiter.WithKeyGetter(keyGenerator(excludedPaths)),
		ginlimiter.WithErrorHandler(func(c *gin.Context, err error) {
			ret := response.ErrorResponseWithCode(xerr.ErrSysInternal)
			c.JSON(http.StatusInternalServerError, ret)
		}),
		ginlimiter.WithLimitReachedHandler(func(c *gin.Context) {
			ret := response.ErrorResponseWithCode(xerr.ErrTooManyRequests)
			c.JSON(http.StatusTooManyRequests, ret)
		}),
		ginlimiter.WithExcludedKey(func(s string) bool {
			if s == RateLimitExcludeKey {
				return true
			}
			return false
		}),
	)
}
