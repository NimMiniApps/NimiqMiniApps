package main

import (
	"context"
	"log/slog"
	"net/http"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const analyticsEventRetention = 180 * 24 * time.Hour

type analyticsMetrics struct {
	Visits           int64 `json:"visits"`
	UniqueVisitors   int64 `json:"unique_visitors"`
	WalletLogins     int64 `json:"wallet_logins"`
	UniqueWallets    int64 `json:"unique_wallets"`
	AppViews         int64 `json:"app_views"`
	UniqueAppViewers int64 `json:"unique_app_viewers"`
	AppOpens         int64 `json:"app_opens"`
	LinkCopies       int64 `json:"link_copies"`
	Activations      int64 `json:"activations"`
	UniqueActivators int64 `json:"unique_activators"`
}

type analyticsDailyPoint struct {
	Day              string `json:"day"`
	Visits           int64  `json:"visits"`
	WalletLogins     int64  `json:"wallet_logins"`
	AppViews         int64  `json:"app_views"`
	AppOpens         int64  `json:"app_opens"`
	LinkCopies       int64  `json:"link_copies"`
	Activations      int64  `json:"activations"`
	UniqueVisitors   int64  `json:"unique_visitors"`
	UniqueAppViewers int64  `json:"unique_app_viewers"`
	UniqueActivators int64  `json:"unique_activators"`
}

type analyticsFunnelStage struct {
	Key           string   `json:"key"`
	Label         string   `json:"label"`
	Count         int64    `json:"count"`
	RateFromPrior *float64 `json:"rate_from_prior,omitempty"`
}

type analyticsAppRow struct {
	Slug             string   `json:"slug"`
	Name             string   `json:"name"`
	Views            int64    `json:"views"`
	UniqueViewers    int64    `json:"unique_viewers"`
	Opens            int64    `json:"opens"`
	LinkCopies       int64    `json:"link_copies"`
	Activations      int64    `json:"activations"`
	UniqueActivators int64    `json:"unique_activators"`
	Conversion       *float64 `json:"conversion"`
}

type adminAnalyticsResponse struct {
	Range             string                 `json:"range"`
	CollectionStarted *time.Time             `json:"collection_started_at"`
	Current           analyticsMetrics       `json:"current"`
	Previous          *analyticsMetrics      `json:"previous,omitempty"`
	Changes           map[string]*float64    `json:"changes,omitempty"`
	Daily             []analyticsDailyPoint  `json:"daily"`
	Funnel            []analyticsFunnelStage `json:"funnel"`
	Apps              []analyticsAppRow      `json:"apps"`
}

type analyticsWindow struct {
	Key   string
	Start time.Time // inclusive
	End   time.Time // exclusive
}

func parseAnalyticsRange(rangeKey string, now time.Time) (current analyticsWindow, previous *analyticsWindow, err error) {
	now = now.UTC()
	end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
	key := rangeKey
	if key == "" {
		key = "30d"
	}
	var days int
	switch key {
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	case "all":
		return analyticsWindow{Key: "all", Start: time.Time{}, End: end}, nil, nil
	default:
		return analyticsWindow{}, nil, errInvalidAnalyticsEvent
	}
	start := end.AddDate(0, 0, -days)
	prevStart := start.AddDate(0, 0, -days)
	prev := analyticsWindow{Key: key, Start: prevStart, End: start}
	return analyticsWindow{Key: key, Start: start, End: end}, &prev, nil
}

func percentChange(current, previous int64) *float64 {
	if previous == 0 {
		return nil
	}
	v := (float64(current) - float64(previous)) / float64(previous) * 100
	return &v
}

func conversionRate(numer, denom int64) *float64 {
	if denom == 0 {
		return nil
	}
	v := float64(numer) / float64(denom) * 100
	return &v
}

func metricsChanges(current, previous analyticsMetrics) map[string]*float64 {
	return map[string]*float64{
		"visits":             percentChange(current.Visits, previous.Visits),
		"unique_visitors":    percentChange(current.UniqueVisitors, previous.UniqueVisitors),
		"wallet_logins":      percentChange(current.WalletLogins, previous.WalletLogins),
		"unique_wallets":     percentChange(current.UniqueWallets, previous.UniqueWallets),
		"app_views":          percentChange(current.AppViews, previous.AppViews),
		"unique_app_viewers": percentChange(current.UniqueAppViewers, previous.UniqueAppViewers),
		"app_opens":          percentChange(current.AppOpens, previous.AppOpens),
		"link_copies":        percentChange(current.LinkCopies, previous.LinkCopies),
		"activations":        percentChange(current.Activations, previous.Activations),
		"unique_activators":  percentChange(current.UniqueActivators, previous.UniqueActivators),
	}
}

func loadCollectionStarted(ctx context.Context, pool *pgxpool.Pool) (*time.Time, error) {
	var started *time.Time
	err := pool.QueryRow(ctx, `SELECT min(occurred_at) FROM analytics_events`).Scan(&started)
	if err != nil {
		return nil, err
	}
	if started == nil {
		err = pool.QueryRow(ctx, `SELECT min(first_seen_at) FROM analytics_uniques`).Scan(&started)
		if err != nil {
			return nil, err
		}
	}
	return started, nil
}

func loadRangeMetrics(ctx context.Context, pool *pgxpool.Pool, start, end time.Time) (analyticsMetrics, error) {
	var m analyticsMetrics
	err := pool.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE event_type = 'catalog_visit'),
			count(DISTINCT visitor_hash) FILTER (WHERE event_type = 'catalog_visit'),
			count(*) FILTER (WHERE event_type = 'wallet_login'),
			count(DISTINCT wallet_hash) FILTER (WHERE event_type = 'wallet_login' AND wallet_hash IS NOT NULL),
			count(*) FILTER (WHERE event_type = 'app_view'),
			count(DISTINCT visitor_hash) FILTER (WHERE event_type = 'app_view'),
			count(*) FILTER (WHERE event_type = 'app_open'),
			count(*) FILTER (WHERE event_type = 'app_link_copy'),
			count(*) FILTER (WHERE event_type IN ('app_open', 'app_link_copy')),
			count(DISTINCT visitor_hash) FILTER (WHERE event_type IN ('app_open', 'app_link_copy'))
		FROM analytics_events
		WHERE occurred_at >= $1 AND occurred_at < $2`, start, end).Scan(
		&m.Visits, &m.UniqueVisitors, &m.WalletLogins, &m.UniqueWallets,
		&m.AppViews, &m.UniqueAppViewers, &m.AppOpens, &m.LinkCopies,
		&m.Activations, &m.UniqueActivators,
	)
	return m, err
}

func loadAllTimeMetrics(ctx context.Context, pool *pgxpool.Pool) (analyticsMetrics, error) {
	var m analyticsMetrics
	err := pool.QueryRow(ctx, `
		SELECT
			coalesce(sum(total_events) FILTER (WHERE event_type = 'catalog_visit' AND app_id IS NULL), 0),
			coalesce(sum(total_events) FILTER (WHERE event_type = 'wallet_login' AND app_id IS NULL), 0),
			coalesce(sum(total_events) FILTER (WHERE event_type = 'app_view'), 0),
			coalesce(sum(total_events) FILTER (WHERE event_type = 'app_open'), 0),
			coalesce(sum(total_events) FILTER (WHERE event_type = 'app_link_copy'), 0)
		FROM analytics_daily`).Scan(&m.Visits, &m.WalletLogins, &m.AppViews, &m.AppOpens, &m.LinkCopies)
	if err != nil {
		return m, err
	}
	m.Activations = m.AppOpens + m.LinkCopies

	err = pool.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE subject_kind = 'visitor' AND event_type = 'catalog_visit' AND app_id IS NULL),
			count(*) FILTER (WHERE subject_kind = 'wallet' AND event_type = 'wallet_login' AND app_id IS NULL),
			count(DISTINCT subject_hash) FILTER (WHERE subject_kind = 'visitor' AND event_type = 'app_view'),
			count(DISTINCT subject_hash) FILTER (WHERE subject_kind = 'visitor' AND event_type IN ('app_open', 'app_link_copy'))
		FROM analytics_uniques`).Scan(&m.UniqueVisitors, &m.UniqueWallets, &m.UniqueAppViewers, &m.UniqueActivators)
	return m, err
}

func loadDailySeries(ctx context.Context, pool *pgxpool.Pool, start, end time.Time) ([]analyticsDailyPoint, error) {
	rows, err := pool.Query(ctx, `
		SELECT day::text, event_type, coalesce(sum(total_events), 0)
		FROM analytics_daily
		WHERE day >= $1::date AND day < $2::date AND (app_id IS NULL OR event_type IN ('app_view', 'app_open', 'app_link_copy'))
		GROUP BY day, event_type
		ORDER BY day`, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byDay := map[string]*analyticsDailyPoint{}
	for rows.Next() {
		var day, eventType string
		var total int64
		if err := rows.Scan(&day, &eventType, &total); err != nil {
			return nil, err
		}
		p := byDay[day]
		if p == nil {
			p = &analyticsDailyPoint{Day: day}
			byDay[day] = p
		}
		switch eventType {
		case "catalog_visit":
			p.Visits += total
		case "wallet_login":
			p.WalletLogins += total
		case "app_view":
			p.AppViews += total
		case "app_open":
			p.AppOpens += total
			p.Activations += total
		case "app_link_copy":
			p.LinkCopies += total
			p.Activations += total
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Unique counts per day from detailed events (available within retention).
	urows, err := pool.Query(ctx, `
		SELECT occurred_at::date::text,
			count(DISTINCT visitor_hash) FILTER (WHERE event_type = 'catalog_visit'),
			count(DISTINCT visitor_hash) FILTER (WHERE event_type = 'app_view'),
			count(DISTINCT visitor_hash) FILTER (WHERE event_type IN ('app_open', 'app_link_copy'))
		FROM analytics_events
		WHERE occurred_at >= $1 AND occurred_at < $2
		GROUP BY occurred_at::date
		ORDER BY 1`, start, end)
	if err != nil {
		return nil, err
	}
	defer urows.Close()
	for urows.Next() {
		var day string
		var uv, uav, ua int64
		if err := urows.Scan(&day, &uv, &uav, &ua); err != nil {
			return nil, err
		}
		p := byDay[day]
		if p == nil {
			p = &analyticsDailyPoint{Day: day}
			byDay[day] = p
		}
		p.UniqueVisitors = uv
		p.UniqueAppViewers = uav
		p.UniqueActivators = ua
	}
	if err := urows.Err(); err != nil {
		return nil, err
	}

	out := make([]analyticsDailyPoint, 0)
	if start.IsZero() {
		days := make([]string, 0, len(byDay))
		for d := range byDay {
			days = append(days, d)
		}
		sort.Strings(days)
		for _, d := range days {
			out = append(out, *byDay[d])
		}
		return out, nil
	}

	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		if p := byDay[key]; p != nil {
			out = append(out, *p)
		} else {
			out = append(out, analyticsDailyPoint{Day: key})
		}
	}
	return out, nil
}

func loadFunnel(ctx context.Context, pool *pgxpool.Pool, start, end time.Time, allTime bool) ([]analyticsFunnelStage, error) {
	var visited, viewed, activated int64
	if allTime {
		err := pool.QueryRow(ctx, `
			SELECT
				count(*) FILTER (WHERE subject_kind = 'visitor' AND event_type = 'catalog_visit' AND app_id IS NULL),
				count(DISTINCT subject_hash) FILTER (WHERE subject_kind = 'visitor' AND event_type = 'app_view'),
				count(DISTINCT subject_hash) FILTER (WHERE subject_kind = 'visitor' AND event_type IN ('app_open', 'app_link_copy'))
			FROM analytics_uniques`).Scan(&visited, &viewed, &activated)
		if err != nil {
			return nil, err
		}
	} else {
		err := pool.QueryRow(ctx, `
			SELECT
				count(DISTINCT visitor_hash) FILTER (WHERE event_type = 'catalog_visit'),
				count(DISTINCT visitor_hash) FILTER (WHERE event_type = 'app_view'),
				count(DISTINCT visitor_hash) FILTER (WHERE event_type IN ('app_open', 'app_link_copy'))
			FROM analytics_events
			WHERE occurred_at >= $1 AND occurred_at < $2`, start, end).Scan(&visited, &viewed, &activated)
		if err != nil {
			return nil, err
		}
	}
	return []analyticsFunnelStage{
		{Key: "visited", Label: "Catalog visitors", Count: visited},
		{Key: "viewed", Label: "App viewers", Count: viewed, RateFromPrior: conversionRate(viewed, visited)},
		{Key: "activated", Label: "Activators", Count: activated, RateFromPrior: conversionRate(activated, viewed)},
	}, nil
}

func loadAppRows(ctx context.Context, pool *pgxpool.Pool, start, end time.Time, allTime bool) ([]analyticsAppRow, error) {
	var query string
	var args []any
	if allTime {
		query = `
			SELECT a.slug, a.name,
				coalesce(sum(d.total_events) FILTER (WHERE d.event_type = 'app_view'), 0),
				coalesce((SELECT count(*) FROM analytics_uniques u
					WHERE u.app_id = a.id AND u.subject_kind = 'visitor' AND u.event_type = 'app_view'), 0),
				coalesce(sum(d.total_events) FILTER (WHERE d.event_type = 'app_open'), 0),
				coalesce(sum(d.total_events) FILTER (WHERE d.event_type = 'app_link_copy'), 0),
				coalesce((SELECT count(DISTINCT subject_hash) FROM analytics_uniques u
					WHERE u.app_id = a.id AND u.subject_kind = 'visitor' AND u.event_type IN ('app_open', 'app_link_copy')), 0)
			FROM apps a
			LEFT JOIN analytics_daily d ON d.app_id = a.id
			GROUP BY a.id, a.slug, a.name
			HAVING coalesce(sum(d.total_events), 0) > 0
			ORDER BY (coalesce(sum(d.total_events) FILTER (WHERE d.event_type IN ('app_open', 'app_link_copy')), 0)) DESC, a.name ASC`
	} else {
		query = `
			SELECT a.slug, a.name,
				count(*) FILTER (WHERE e.event_type = 'app_view'),
				count(DISTINCT e.visitor_hash) FILTER (WHERE e.event_type = 'app_view'),
				count(*) FILTER (WHERE e.event_type = 'app_open'),
				count(*) FILTER (WHERE e.event_type = 'app_link_copy'),
				count(DISTINCT e.visitor_hash) FILTER (WHERE e.event_type IN ('app_open', 'app_link_copy'))
			FROM apps a
			JOIN analytics_events e ON e.app_id = a.id
			WHERE e.occurred_at >= $1 AND e.occurred_at < $2
			GROUP BY a.id, a.slug, a.name
			ORDER BY (count(*) FILTER (WHERE e.event_type IN ('app_open', 'app_link_copy'))) DESC, a.name ASC`
		args = []any{start, end}
	}
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]analyticsAppRow, 0)
	for rows.Next() {
		var row analyticsAppRow
		if err := rows.Scan(&row.Slug, &row.Name, &row.Views, &row.UniqueViewers, &row.Opens, &row.LinkCopies, &row.UniqueActivators); err != nil {
			return nil, err
		}
		row.Activations = row.Opens + row.LinkCopies
		row.Conversion = conversionRate(row.UniqueActivators, row.UniqueViewers)
		out = append(out, row)
	}
	return out, rows.Err()
}

func buildAdminAnalytics(ctx context.Context, pool *pgxpool.Pool, rangeKey string, now time.Time) (adminAnalyticsResponse, error) {
	current, previous, err := parseAnalyticsRange(rangeKey, now)
	if err != nil {
		return adminAnalyticsResponse{}, err
	}

	started, err := loadCollectionStarted(ctx, pool)
	if err != nil {
		return adminAnalyticsResponse{}, err
	}

	resp := adminAnalyticsResponse{
		Range:             current.Key,
		CollectionStarted: started,
		Daily:             []analyticsDailyPoint{},
		Funnel:            []analyticsFunnelStage{},
		Apps:              []analyticsAppRow{},
	}

	allTime := current.Key == "all"
	if allTime {
		resp.Current, err = loadAllTimeMetrics(ctx, pool)
		if err != nil {
			return resp, err
		}
		resp.Daily, err = loadDailySeries(ctx, pool, time.Time{}, current.End)
		if err != nil {
			return resp, err
		}
	} else {
		resp.Current, err = loadRangeMetrics(ctx, pool, current.Start, current.End)
		if err != nil {
			return resp, err
		}
		if previous != nil {
			prev, err := loadRangeMetrics(ctx, pool, previous.Start, previous.End)
			if err != nil {
				return resp, err
			}
			resp.Previous = &prev
			resp.Changes = metricsChanges(resp.Current, prev)
		}
		resp.Daily, err = loadDailySeries(ctx, pool, current.Start, current.End)
		if err != nil {
			return resp, err
		}
	}

	resp.Funnel, err = loadFunnel(ctx, pool, current.Start, current.End, allTime)
	if err != nil {
		return resp, err
	}
	resp.Apps, err = loadAppRows(ctx, pool, current.Start, current.End, allTime)
	return resp, err
}

func (s *server) adminAnalytics(w http.ResponseWriter, r *http.Request) {
	rangeKey := r.URL.Query().Get("range")
	resp, err := buildAdminAnalytics(r.Context(), s.pool, rangeKey, time.Now().UTC())
	if err == errInvalidAnalyticsEvent {
		writeError(w, http.StatusBadRequest, "invalid range")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func pruneAnalyticsEvents(ctx context.Context, pool *pgxpool.Pool, now time.Time) error {
	cutoff := now.UTC().Add(-analyticsEventRetention)
	_, err := pool.Exec(ctx, `DELETE FROM analytics_events WHERE occurred_at < $1`, cutoff)
	return err
}

func (s *server) startAnalyticsRetentionWorker(ctx context.Context) {
	go func() {
		run := func() {
			if err := pruneAnalyticsEvents(ctx, s.pool, time.Now().UTC()); err != nil {
				slog.Error("analytics retention prune failed", "error", err.Error())
			}
		}
		run()
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}
