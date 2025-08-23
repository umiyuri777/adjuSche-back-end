package application

import (
	"adjuSche-back-end/repository"
	"adjuSche-back-end/servise"
	"context"
	"fmt"
	"time"
)

type InviteSummary struct {
	EventName   string
	VotedCount  int
	Memo        string
	PeriodStart time.Time
	PeriodEnd   time.Time
	DurationMin int
}

type PossibleSlot struct {
	ID                   int
	Date                 string
	PeriodStart          time.Time
	PeriodEnd            time.Time
	ParticipateMemberNum int
}

// BuildInviteResponse はイベントIDとGoogleトークンから空き時間候補を構築する
func BuildInviteResponse(ctx context.Context, eventID int64, tokenString string, credFile string) (InviteSummary, []PossibleSlot, error) {
	repo, err := repository.NewSupabaseRepository()
	if err != nil {
		return InviteSummary{}, nil, fmt.Errorf("failed to init repository: %w", err)
	}

	ev, err := repo.GetEventByID(ctx, eventID)
	if err != nil {
		return InviteSummary{}, nil, err
	}
	cond, err := repo.GetEventConditionByEventID(ctx, eventID)
	if err != nil {
		return InviteSummary{}, nil, err
	}

	// Google カレンダーから空き時間抽出
	cal, err := servise.NewCalendarServiceFromTokenString(tokenString, credFile)
	if err != nil {
		return InviteSummary{}, nil, fmt.Errorf("failed to init calendar service: %w", err)
	}

	free, err := cal.GetFreeIntervalsInRange(cond.PeriodStart, cond.PeriodEnd, cond.DurationMin)
	if err != nil {
		return InviteSummary{}, nil, err
	}

	// ユニーク投票者数（Availabilities に提出済みのユーザー）
	voted, _ := repo.CountDistinctAvailabilityUsersByEventID(ctx, eventID)

	// レスポンス候補の整形
	slots := make([]PossibleSlot, 0, len(free))
	for i, iv := range free {
		dateStr := iv.Start.Format("2006-01-02")
		slots = append(slots, PossibleSlot{
			ID:                   i + 1,
			Date:                 dateStr,
			PeriodStart:          iv.Start,
			PeriodEnd:            iv.End,
			ParticipateMemberNum: 0,
		})
	}

	memo := ""
	if ev.Note.Valid {
		memo = ev.Note.String
	}

	summary := InviteSummary{
		EventName:   ev.Title,
		VotedCount:  voted,
		Memo:        memo,
		PeriodStart: cond.PeriodStart,
		PeriodEnd:   cond.PeriodEnd,
		DurationMin: cond.DurationMin,
	}

	return summary, slots, nil
}
