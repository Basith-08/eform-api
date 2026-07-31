package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"eform/backend/internal/domain"
	"eform/backend/pkg/auth"
	"eform/backend/pkg/response"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

const ContextClaimsKey = "claims"

func CORSMiddleware(frontendURL string) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data: http: https:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self' http: https:")
		c.Next()
	}
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func RateLimiter(rps int, burst int) gin.HandlerFunc {
	var (
		mu       sync.Mutex
		visitors = map[string]*visitor{}
	)

	go func() {
		for range time.Tick(time.Minute) {
			mu.Lock()
			for ip, entry := range visitors {
				if time.Since(entry.lastSeen) > 3*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		entry, ok := visitors[ip]
		if !ok {
			entry = &visitor{
				limiter: rate.NewLimiter(rate.Limit(rps), burst),
			}
			visitors[ip] = entry
		}
		entry.lastSeen = time.Now()
		mu.Unlock()

		if !entry.limiter.Allow() {
			response.Error(c, http.StatusTooManyRequests, "too many requests")
			c.Abort()
			return
		}

		c.Next()
	}
}

func AuthRequired(manager *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, "missing bearer token")
			c.Abort()
			return
		}

		claims, err := manager.Parse(strings.TrimPrefix(header, "Bearer "))
		if err != nil || claims.Type != "access" {
			response.Error(c, http.StatusUnauthorized, "invalid access token")
			c.Abort()
			return
		}

		parsedID, err := uuid.Parse(claims.UserID)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid access token")
			c.Abort()
			return
		}

		c.Set(ContextClaimsKey, domain.AuthClaims{
			UserID: parsedID,
			Role:   claims.Role,
			Email:  claims.Email,
		})

		c.Next()
	}
}

func RoleRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetClaims(c)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "missing access token")
			c.Abort()
			return
		}

		for _, role := range roles {
			if claims.Role == role {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, "insufficient permissions")
		c.Abort()
	}
}

func GetClaims(c *gin.Context) (domain.AuthClaims, bool) {
	value, ok := c.Get(ContextClaimsKey)
	if !ok {
		return domain.AuthClaims{}, false
	}

	claims, ok := value.(domain.AuthClaims)
	return claims, ok
}
