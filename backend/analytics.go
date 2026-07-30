package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"
)

type analyticsEventType string

const (
	analyticsCatalogVisit analyticsEventType = "catalog_visit"
	analyticsWalletLogin  analyticsEventType = "wallet_login"
	analyticsAppView      analyticsEventType = "app_view"
	analyticsAppOpen      analyticsEventType = "app_open"
	analyticsAppLinkCopy  analyticsEventType = "app_link_copy"
)

type analyticsEventInput struct {
	EventType     analyticsEventType
	VisitorID     string
	SessionID     string
	WalletAddress string
	AppID         string
	OccurredAt    time.Time
}

var (
	errInvalidAnalyticsIdentifier = errors.New("invalid analytics identifier")
	errInvalidAnalyticsEvent      = errors.New("invalid analytics event")
	errAnalyticsHashSecretUnset   = errors.New("ANALYTICS_HASH_SECRET is empty; analytics recording is disabled")
)

const maxAnalyticsIdentifierLen = 64

// uuidShapedRe matches the client-generated, opaque UUID-style identifiers
// used for visitor/session tracking (e.g. crypto.randomUUID()).
var uuidShapedRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// appAssociatedEventTypes require a valid app_id; the rest must not carry one.
var appAssociatedEventTypes = map[analyticsEventType]bool{
	analyticsAppView:     true,
	analyticsAppOpen:     true,
	analyticsAppLinkCopy: true,
}

func validateAnalyticsIdentifier(value string) error {
	if value == "" || len(value) > maxAnalyticsIdentifierLen || !uuidShapedRe.MatchString(value) {
		return errInvalidAnalyticsIdentifier
	}
	return nil
}

// hashAnalyticsIdentifier derives a stable, one-way hash for a raw browser or
// wallet identifier. Namespacing the HMAC message with kind keeps visitor,
// session, and wallet hashes from ever colliding with each other.
func hashAnalyticsIdentifier(secret, kind, value string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(kind + "\x00" + value))
	return hex.EncodeToString(mac.Sum(nil))
}

// analyticsHashSecret returns the configured ANALYTICS_HASH_SECRET, or an
// error if it is unset. There is no insecure default: like
// WALLET_AUTH_SECRET, an empty secret means the feature is unavailable
// rather than silently hashing with a known value.
func analyticsHashSecret() (string, error) {
	secret := env("ANALYTICS_HASH_SECRET", "")
	if secret == "" {
		return "", errAnalyticsHashSecretUnset
	}
	return secret, nil
}

// recordAnalyticsEvent validates and persists one analytics event, updating
// the daily aggregate, uniqueness, and (for app_view/app_open) the legacy
// app_stats_daily counters in a single transaction. It returns recorded=false
// when the event was a deduplicated catalog visit or app view.
func (s *server) recordAnalyticsEvent(ctx context.Context, input analyticsEventInput) (bool, error) {
	switch input.EventType {
	case analyticsCatalogVisit, analyticsWalletLogin, analyticsAppView, analyticsAppOpen, analyticsAppLinkCopy:
	default:
		return false, errInvalidAnalyticsEvent
	}

	if err := validateAnalyticsIdentifier(input.VisitorID); err != nil {
		return false, err
	}
	if err := validateAnalyticsIdentifier(input.SessionID); err != nil {
		return false, err
	}

	requiresApp := appAssociatedEventTypes[input.EventType]
	if requiresApp && input.AppID == "" {
		return false, errInvalidAnalyticsEvent
	}
	if !requiresApp && input.AppID != "" {
		return false, errInvalidAnalyticsEvent
	}

	// Only the server-side wallet-login path may attach a wallet address;
	// public event submissions can never carry one.
	if input.WalletAddress != "" && input.EventType != analyticsWalletLogin {
		return false, errInvalidAnalyticsEvent
	}

	secret, err := analyticsHashSecret()
	if err != nil {
		return false, err
	}

	occurredAt := input.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	visitorHash := hashAnalyticsIdentifier(secret, "visitor", input.VisitorID)
	sessionHash := hashAnalyticsIdentifier(secret, "session", input.SessionID)

	var walletHash *string
	if input.WalletAddress != "" {
		h := hashAnalyticsIdentifier(secret, "wallet", input.WalletAddress)
		walletHash = &h
	}
	var appID *string
	if input.AppID != "" {
		appID = &input.AppID
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		INSERT INTO analytics_events (occurred_at, event_type, visitor_hash, session_hash, wallet_hash, app_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING`,
		occurredAt, string(input.EventType), visitorHash, sessionHash, walletHash, appID)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() == 0 {
		return false, tx.Commit(ctx)
	}

	day := occurredAt.UTC().Format("2006-01-02")

	if _, err := tx.Exec(ctx, `
		INSERT INTO analytics_daily (day, event_type, app_id, total_events)
		VALUES ($1::date, $2, $3, 1)
		ON CONFLICT (day, event_type, app_id) DO UPDATE SET total_events = analytics_daily.total_events + 1`,
		day, string(input.EventType), appID); err != nil {
		return false, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO analytics_uniques (subject_kind, event_type, app_id, subject_hash, first_seen_at, last_seen_at)
		VALUES ('visitor', $1, $2, $3, $4, $4)
		ON CONFLICT (subject_kind, event_type, app_id, subject_hash) DO UPDATE
		SET last_seen_at = GREATEST(analytics_uniques.last_seen_at, EXCLUDED.last_seen_at)`,
		string(input.EventType), appID, visitorHash, occurredAt); err != nil {
		return false, err
	}

	if walletHash != nil {
		if _, err := tx.Exec(ctx, `
			INSERT INTO analytics_uniques (subject_kind, event_type, app_id, subject_hash, first_seen_at, last_seen_at)
			VALUES ('wallet', $1, $2, $3, $4, $4)
			ON CONFLICT (subject_kind, event_type, app_id, subject_hash) DO UPDATE
			SET last_seen_at = GREATEST(analytics_uniques.last_seen_at, EXCLUDED.last_seen_at)`,
			string(input.EventType), appID, *walletHash, occurredAt); err != nil {
			return false, err
		}
	}

	if input.EventType == analyticsAppView || input.EventType == analyticsAppOpen {
		opens, views := 0, 0
		if input.EventType == analyticsAppView {
			views = 1
		} else {
			opens = 1
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO app_stats_daily (app_id, day, opens, views)
			VALUES ($1, $2::date, $3, $4)
			ON CONFLICT (app_id, day) DO UPDATE SET
				opens = app_stats_daily.opens + $3,
				views = app_stats_daily.views + $4`,
			*appID, day, opens, views); err != nil {
			return false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}

// analyticsTrackRequest is the strict public shape accepted by
// POST /api/analytics/events. Unknown fields (including any wallet_*
// field) are rejected outright by trackAnalyticsEvent rather than ignored,
// so the public endpoint can never smuggle in wallet data.
type analyticsTrackRequest struct {
	Event     string `json:"event"`
	VisitorID string `json:"visitor_id"`
	SessionID string `json:"session_id"`
	AppSlug   string `json:"app_slug"`
}

// publicAnalyticsEventTypes are the only events a client may report directly;
// wallet_login is server-recorded only, from a verified auth session.
var publicAnalyticsEventTypes = map[analyticsEventType]bool{
	analyticsCatalogVisit: true,
	analyticsAppView:      true,
	analyticsAppOpen:      true,
	analyticsAppLinkCopy:  true,
}

// trackAnalyticsEvent ingests one anonymous, client-reported product event.
// It never accepts or persists a wallet identifier or the caller's IP; the IP
// is used only as part of the in-memory rate-limit key.
func (s *server) trackAnalyticsEvent(w http.ResponseWriter, r *http.Request) {
	if s.analyticsHashSecret == "" {
		writeError(w, http.StatusServiceUnavailable, "analytics is not configured")
		return
	}

	var req analyticsTrackRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	eventType := analyticsEventType(req.Event)
	if !publicAnalyticsEventTypes[eventType] {
		writeError(w, http.StatusBadRequest, "unknown or unsupported event")
		return
	}
	if err := validateAnalyticsIdentifier(req.VisitorID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateAnalyticsIdentifier(req.SessionID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	limiterKey := clientIP(r) + ":" + req.VisitorID
	if s.analyticsLimiter != nil && !s.analyticsLimiter.allow(limiterKey) {
		writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	input := analyticsEventInput{
		EventType: eventType,
		VisitorID: req.VisitorID,
		SessionID: req.SessionID,
	}
	if appAssociatedEventTypes[eventType] {
		if req.AppSlug == "" {
			writeError(w, http.StatusBadRequest, "app_slug is required for this event")
			return
		}
		appID, err := s.appIDForSlug(r.Context(), req.AppSlug)
		if err != nil {
			writeError(w, http.StatusBadRequest, "unknown app_slug")
			return
		}
		input.AppID = appID
	} else if req.AppSlug != "" {
		writeError(w, http.StatusBadRequest, "app_slug is not valid for this event")
		return
	}

	if _, err := s.recordAnalyticsEvent(r.Context(), input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
