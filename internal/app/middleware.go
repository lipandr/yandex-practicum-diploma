package app

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/lipandr/yandex-practicum-diploma/internal/service"
	"github.com/lipandr/yandex-practicum-diploma/internal/types"
)

const authorizationScheme = "Bearer"

var noAuth = map[string]interface{}{
	"/api/user/register": nil,
	"/api/user/login":    nil,
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// GzipMiddleware middleware метод обрабатывающий сжатие gzip.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer func() {
				_ = gz.Close()
			}()

			body, err := io.ReadAll(gz)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = ioutil.NopCloser(bytes.NewReader(body))
			r.ContentLength = int64(len(body))
		}
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			if _, err := io.WriteString(w, err.Error()); err != nil {
				return
			}
			return
		}
		defer func() {
			_ = gz.Close()
		}()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

// AuthMiddleware middleware метода аутентификации/авторизации пользователя.
func AuthMiddleware(svc service.Service) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := noAuth[r.RequestURI]; ok {
				next.ServeHTTP(w, r)
				return
			}
			token, err := getTokenFromAuthHeader(r.Header.Get("Authorization"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			id, err := svc.GetUserIDByToken(token)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), types.UserID, id)
			req := r.WithContext(ctx)
			next.ServeHTTP(w, req)
		})
	}
}

// Метод-helper AuthMiddleware получения токена из заголовка запроса.
func getTokenFromAuthHeader(headerVal string) (string, error) {
	var token string

	if strings.HasPrefix(headerVal, authorizationScheme) {
		headerParts := strings.Split(headerVal, " ")
		if len(headerParts) != 2 {
			return "", errors.New("wrong auth header")
		}
		token = headerParts[1]
	}
	return token, nil
}
