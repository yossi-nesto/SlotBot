package slack

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

func VerifySignature(signingSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timestamp := r.Header.Get("X-Slack-Request-Timestamp")
			if timestamp == "" {
				slog.Warn("Slack signature verification failed: missing timestamp")
				http.Error(w, "Missing timestamp", http.StatusUnauthorized)
				return
			}

			// Check if timestamp is too old (replay attack)
			tsInt, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				slog.Warn("Slack signature verification failed: invalid timestamp", "error", err)
				http.Error(w, "Invalid timestamp", http.StatusUnauthorized)
				return
			}
			if time.Now().Unix()-tsInt > 60*5 {
				slog.Warn("Slack signature verification failed: timestamp too old")
				http.Error(w, "Timestamp too old", http.StatusUnauthorized)
				return
			}

			signature := r.Header.Get("X-Slack-Signature")
			if signature == "" {
				slog.Warn("Slack signature verification failed: missing signature")
				http.Error(w, "Missing signature", http.StatusUnauthorized)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				slog.Error("Failed to read request body", "error", err)
				http.Error(w, "Failed to read body", http.StatusInternalServerError)
				return
			}
			// Restore body for next handler
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			sigBase := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
			mac := hmac.New(sha256.New, []byte(signingSecret))
			mac.Write([]byte(sigBase))
			expectedSig := "v0=" + hex.EncodeToString(mac.Sum(nil))

			if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
				slog.Warn("Slack signature verification failed: signature mismatch",
					"received", signature[:20]+"...",
					"expected", expectedSig[:20]+"...",
					"signing_secret_length", len(signingSecret))
				http.Error(w, "Invalid signature", http.StatusUnauthorized)
				return
			}

			slog.Debug("Slack signature verified successfully")
			next.ServeHTTP(w, r)
		})
	}
}
