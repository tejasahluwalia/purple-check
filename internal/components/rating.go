package components

import (
	"context"
	"log"

	"purple-check/internal/database"
)

func GetProfileRating(ctx context.Context, username string) (float64, int, error) {
	if _, err := database.PullDB(ctx); err != nil {
		log.Println("Error pulling database changes.", err)
	}

	db, closer, err := database.GetDB(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer closer()

	var positiveRatings int
	var totalRatings int

	err = db.QueryRow("SELECT COUNT(*) FROM feedback WHERE receiver = ? AND rating = 'POSITIVE'", username).Scan(&positiveRatings)
	if err != nil {
		return 0, 0, err
	}

	err = db.QueryRow("SELECT COUNT(*) FROM feedback WHERE receiver = ?", username).Scan(&totalRatings)
	if err != nil {
		return 0, 0, err
	}

	return float64(positiveRatings), totalRatings, nil
}
