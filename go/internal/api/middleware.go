package api

import (
	"log"
	"net/http"
	"time"
)

const (
	corsAllowOrigin      = "Access-Control-Allow-Origin"
	corsAllowMethods     = "Access-Control-Allow-Methods"
	corsAllowHeaders     = "Access-Control-Allow-Headers"
	corsAllowCredentials = "Access-Control-Allow-Credentials"
	allowedOrigin        = "http://localhost:5173"
	allowedMethods       = "GET, POST, PUT, DELETE, OPTIONS"
	allowedHeaders       = "Content-Type, Authorization"
	allowedCredentials   = "true"
	internalServerError  = "Internal server error"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf(
			"%s %s %s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				http.Error(w, internalServerError, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(corsAllowOrigin, allowedOrigin)
		w.Header().Set(corsAllowMethods, allowedMethods)
		w.Header().Set(corsAllowHeaders, allowedHeaders)
		w.Header().Set(corsAllowCredentials, allowedCredentials)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
