package application

import (
	"adjuSche-back-end/repository"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// CreateEventInput はイベント作成に必要な入力を表す
type CreateEventInput struct {
	HostUserID       string
	Title            string
	Memo             string
	ParticipantCount int
	PeriodStart      string
	PeriodEnd        string
	TimeStart        string
	TimeEnd          string
	DurationMin      int
}

// CreateEventAndCondition は Events と EventConditions を作成し、作成したイベントIDを返す
func CreateEventAndCondition(ctx context.Context, in CreateEventInput) (int64, error) {
	repo, err := repository.NewSupabaseRepository()
	if err != nil {
		return 0, fmt.Errorf("failed to init repository: %w", err)
	}

	now := time.Now()

	ev := &repository.Events{
		HostUserID:       in.HostUserID,
		Title:            in.Title,
		Note:             sql.NullString{String: in.Memo, Valid: in.Memo != ""},
		ParticipantCount: int64(in.ParticipantCount),
		Status:           repository.EventStatusDraft,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := repo.CreateEvent(ctx, ev); err != nil {
		return 0, err
	}

	// 期間のパース（RFC3339 もしくは日付のみ 2006-01-02 を許容）
	ps, err := parseRFC3339OrDate(in.PeriodStart)
	if err != nil {
		return 0, fmt.Errorf("invalid periodStart: %w", err)
	}
	pe, err := parseRFC3339OrDate(in.PeriodEnd)
	if err != nil {
		return 0, fmt.Errorf("invalid periodEnd: %w", err)
	}

	// time_type 判定: デフォルトは all_day(4)、開始/終了が指定されれば custom(3)
	timeType := 4
	var tStart, tEnd sql.NullString
	if in.TimeStart != "" || in.TimeEnd != "" {
		timeType = 3
		if in.TimeStart != "" {
			tStart = sql.NullString{String: in.TimeStart, Valid: true}
		}
		if in.TimeEnd != "" {
			tEnd = sql.NullString{String: in.TimeEnd, Valid: true}
		}
	}

	cond := &repository.EventCondition{
		EventID:     ev.ID,
		PeriodStart: ps,
		PeriodEnd:   pe,
		TimeType:    timeType,
		TimeStart:   tStart,
		TimeEnd:     tEnd,
		DurationMin: in.DurationMin,
		CreatedAt:   now,
	}
	if err := repo.CreateEventCondition(ctx, cond); err != nil {
		return 0, err
	}

	return ev.ID, nil
}

func parseRFC3339OrDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}
	if len(value) >= 10 && value[4] == '-' && value[7] == '-' && len(value) == len("2006-01-02") {
		return time.Parse("2006-01-02", value)
	}
	return time.Parse(time.RFC3339, value)
}
