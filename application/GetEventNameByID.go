package application

import (
	"context"
	"adjuSche-back-end/repository"
)

func GetEventNameByID(ctx context.Context, eventID int64) (string, error) {
	repo, err := repository.NewSupabaseRepository()
	if err != nil {
		return "", err
	}

	event, err := repo.GetEventByID(ctx, eventID)
	if err != nil {
		return "", err
	}

	return event.Title, nil
}