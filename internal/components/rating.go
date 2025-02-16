package components

import (
	"database/sql"
	"log"

	"purple-check/internal/database"
)

func GetProfileRating(username string) (float64, int) {
	db, closer := database.GetDB()
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
