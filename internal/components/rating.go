package components

import (
	"context"
	"database/sql"
	"log"

	"purple-check/internal/database"
)

func GetProfileRating(ctx context.Context, username string) (float64, int) {
	if _, err := database.PullDB(ctx); err != nil {
		log.Println("Error pulling database changes.", err)
	}

	db, closer := database.GetDB(ctx)
	defer closer()

	var rating sql.NullFloat64
	var totalRatings int

	err := db.QueryRow("SELECT COUNT(*) FROM feedback WHERE receiver = ? AND rating = 'POSITIVE'", username).Scan(&rating)
	if err != nil {
		log.Fatal("Error querying database.", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM feedback WHERE receiver = ?", username).Scan(&totalRatings)
	if err != nil {
		log.Fatal("Error querying database.", err)
	}

	return rating.Float64, totalRatings
}
