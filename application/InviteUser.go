package application

import (
	"adjuSche-back-end/repository"
	"adjuSche-back-end/servise"
	"context"
	"fmt"
	"sort"
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

// TimeSlot は時間スロットを表す構造体
type TimeSlot struct {
	Start time.Time
	End   time.Time
}

// calculateOverlappingSlots は全参加者の空き時間の重複期間を計算します
func calculateOverlappingSlots(allAvailabilities []repository.Availability, newUserSlots []servise.TimeInterval, durationMin int) []PossibleSlot {
	// 全てのユーザーの空き時間を TimeSlot に変換
	userSlots := make(map[string][]TimeSlot)

	// 既存参加者の空き時間を追加
	for _, av := range allAvailabilities {
		start, err := time.Parse(time.RFC3339, av.AvailableStart)
		if err != nil {
			continue
		}
		end, err := time.Parse(time.RFC3339, av.AvailableEnd)
		if err != nil {
			continue
		}

		if userSlots[av.UserID] == nil {
			userSlots[av.UserID] = make([]TimeSlot, 0)
		}
		userSlots[av.UserID] = append(userSlots[av.UserID], TimeSlot{Start: start, End: end})
	}

	// 新しいユーザーの空き時間を追加（仮のユーザーIDを使用）
	newUserID := "new_user"
	for _, interval := range newUserSlots {
		if userSlots[newUserID] == nil {
			userSlots[newUserID] = make([]TimeSlot, 0)
		}
		userSlots[newUserID] = append(userSlots[newUserID], TimeSlot{Start: interval.Start, End: interval.End})
	}

	// 参加者数を計算
	participantCount := len(userSlots)
	if participantCount == 0 {
		return []PossibleSlot{}
	}

	// 全ユーザーの重複期間を計算
	overlapping := findOverlappingTimeSlots(userSlots, durationMin)

	// PossibleSlot に変換
	slots := make([]PossibleSlot, 0, len(overlapping))
	for i, slot := range overlapping {
		dateStr := slot.Start.Format("2006-01-02")
		slots = append(slots, PossibleSlot{
			ID:                   i + 1,
			Date:                 dateStr,
			PeriodStart:          slot.Start,
			PeriodEnd:            slot.End,
			ParticipateMemberNum: participantCount,
		})
	}

	return slots
}

// findOverlappingTimeSlots は全ユーザーの空き時間の重複期間を見つけます
func findOverlappingTimeSlots(userSlots map[string][]TimeSlot, durationMin int) []TimeSlot {
	if len(userSlots) == 0 {
		return []TimeSlot{}
	}

	// 全てのイベント（開始・終了）を収集
	type Event struct {
		Time    time.Time
		IsStart bool
		UserID  string
	}

	var events []Event
	for userID, slots := range userSlots {
		for _, slot := range slots {
			events = append(events, Event{Time: slot.Start, IsStart: true, UserID: userID})
			events = append(events, Event{Time: slot.End, IsStart: false, UserID: userID})
		}
	}

	// 時間順にソート
	sort.Slice(events, func(i, j int) bool {
		if events[i].Time.Equal(events[j].Time) {
			// 同じ時刻の場合、終了イベントを先に処理
			return !events[i].IsStart && events[j].IsStart
		}
		return events[i].Time.Before(events[j].Time)
	})

	// アクティブなユーザーを追跡
	activeUsers := make(map[string]bool)
	var result []TimeSlot
	var currentStart time.Time
	totalUsers := len(userSlots)

	for _, event := range events {
		currentActiveCount := len(activeUsers)

		// 全ユーザーがアクティブな期間を記録
		if currentActiveCount == totalUsers && !currentStart.IsZero() {
			duration := event.Time.Sub(currentStart)
			if duration >= time.Duration(durationMin)*time.Minute {
				result = append(result, TimeSlot{Start: currentStart, End: event.Time})
			}
		}

		// アクティブユーザーリストを更新
		if event.IsStart {
			activeUsers[event.UserID] = true
		} else {
			delete(activeUsers, event.UserID)
		}

		// 新しい期間の開始を記録
		if len(activeUsers) == totalUsers {
			currentStart = event.Time
		} else {
			currentStart = time.Time{} // リセット
		}
	}

	return result
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

	// 既存参加者の空き時間を取得
	allAvailabilities, err := repo.ListAvailabilitiesByEventID(ctx, eventID)
	if err != nil {
		fmt.Printf("ListAvailabilitiesByEventID エラー: %v\n", err)
		return InviteSummary{}, nil, err
	}
	fmt.Printf("既存参加者の空き時間レコード数: %d\n", len(allAvailabilities))

	// ユニーク投票者数（Availabilities に提出済みのユーザー + 新規ユーザー）
	voted, _ := repo.CountDistinctAvailabilityUsersByEventID(ctx, eventID)
	// 新規ユーザーが追加されるため +1
	voted = voted + 1

	// 全員（既存 + 新規）の空き時間の重複期間を計算
	slots := calculateOverlappingSlots(allAvailabilities, free, cond.DurationMin)
	fmt.Printf("計算された重複期間数: %d\n", len(slots))

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

// RegisterEventParticipant はユーザーをイベントの参加者として登録します
func RegisterEventParticipant(ctx context.Context, eventID int64, userID string) error {
	fmt.Printf("RegisterEventParticipant: eventID=%d, userID=%s\n", eventID, userID)

	repo, err := repository.NewSupabaseRepository()
	if err != nil {
		return fmt.Errorf("failed to init repository: %w", err)
	}

	participant, err := repo.GetOrCreateEventParticipant(ctx, eventID, userID)
	if err != nil {
		return fmt.Errorf("failed to register event participant: %w", err)
	}

	fmt.Printf("参加者登録完了: participantID=%d, status=%d\n", participant.ID, participant.Status)
	return nil
}
