package middleware

import (
	"net"
	"net/http"

	"go.uber.org/zap"
)

func TrustedSubnet(trustedSubnet string, logger *zap.Logger) func(next http.Handler) http.Handler {
	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if trustedSubnet != "" && err != nil {
		logger.Warn("invalid trusted subnet, denying all requests to internal endpoints",
			zap.String("trusted_subnet", trustedSubnet),
			zap.Error(err))
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ipNet == nil {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			ipStr := r.Header.Get("X-Real-IP")
			ip := net.ParseIP(ipStr)
			if ip == nil || !ipNet.Contains(ip) {
				logger.Warn("request from untrusted IP denied",
					zap.String("x_real_ip", ipStr),
					zap.String("trusted_subnet", trustedSubnet),
					zap.String("uri", r.RequestURI))
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
