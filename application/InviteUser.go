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
	fmt.Printf("BuildInviteResponse: eventID=%d を開始します\n", eventID)

	repo, err := repository.NewSupabaseRepository()
	if err != nil {
		fmt.Printf("repository初期化エラー: %v\n", err)
		return InviteSummary{}, nil, fmt.Errorf("failed to init repository: %w", err)
	}

	fmt.Printf("GetEventByID を呼び出します: eventID=%d\n", eventID)
	ev, err := repo.GetEventByID(ctx, eventID)
	if err != nil {
		fmt.Printf("GetEventByID エラー: %v\n", err)
		return InviteSummary{}, nil, err
	}
	fmt.Printf("GetEventByID 成功: title=%s\n", ev.Title)

	fmt.Printf("GetEventConditionByEventID を呼び出します: eventID=%d\n", eventID)
	cond, err := repo.GetEventConditionByEventID(ctx, eventID)
	if err != nil {
		fmt.Printf("GetEventConditionByEventID エラー: %v\n", err)
		return InviteSummary{}, nil, err
	}
	fmt.Printf("GetEventConditionByEventID 成功: period=%s to %s\n", cond.PeriodStart.Format("2006-01-02"), cond.PeriodEnd.Format("2006-01-02"))

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

// SaveUserAvailabilitiesFromCalendar は、与えられた空き時間を Availabilities に保存する
// available_start/end は RFC3339 の時刻文字列、available_date は YYYY-MM-DD
func SaveUserAvailabilitiesFromCalendar(ctx context.Context, eventID int64, userID string, intervals []servise.TimeInterval) error {
	fmt.Printf("SaveUserAvailabilitiesFromCalendar: eventID=%d, userID=%s, intervals=%d\n", eventID, userID, len(intervals))

	repo, err := repository.NewSupabaseRepository()
	if err != nil {
		return fmt.Errorf("failed to init repository: %w", err)
	}

	avs := make([]repository.Availability, 0, len(intervals))
	now := time.Now()
	for i, iv := range intervals {
		dateStr := iv.Start.Format("2006-01-02")
		startStr := iv.Start.Format(time.RFC3339)
		endStr := iv.End.Format(time.RFC3339)
		av := repository.Availability{
			EventID:        eventID,
			UserID:         userID,
			AvailableDate:  dateStr,
			AvailableStart: startStr,
			AvailableEnd:   endStr,
			Sourse:         0, // 0: google_calendar
			CreatedAt:      now,
		}
		avs = append(avs, av)
		fmt.Printf("  [%d] %s: %s - %s\n", i, dateStr, startStr, endStr)
	}

	fmt.Printf("ReplaceUserAvailabilitiesForEvent を呼び出します\n")
	if err := repo.ReplaceUserAvailabilitiesForEvent(ctx, eventID, userID, avs); err != nil {
		fmt.Printf("ReplaceUserAvailabilitiesForEvent エラー: %v\n", err)
		return err
	}
	fmt.Printf("ReplaceUserAvailabilitiesForEvent 完了\n")
	return nil
}
