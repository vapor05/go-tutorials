package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type Album struct {
	ID     int64
	Title  string
	Artist string
	Price  float32
}

var db *sql.DB

func main() {
	db_pass := os.Getenv("DB_PASS")
	db_url := "postgres://postgres:" + db_pass + "@localhost:5432/recordings"
	var err error
	db, err = sql.Open("pgx", db_url)
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("connected!")

	albums, err := albumsByArtist("John Coltrane")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Albums found: %v\n", albums)

	// Hard-code ID 2 here to test the query
	alb, err := albumByID(2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Album found: %v\n", alb)

	// add an album
	albID, err := addAlbum(Album{
		Title:  "The Modern Sound of Betty Carter",
		Artist: "Betty Carter",
		Price:  49.99,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ID of added album: %v\n", albID)
}

// queries for albums that have the specified artist name
func albumsByArtist(name string) ([]Album, error) {
	// An albums slice to hold data from returned rows
	var albums []Album

	rows, err := db.Query("SELECT * FROM album WHERE artist = $1", name)
	if err != nil {
		return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
	}

	defer rows.Close()

	// loop through rows, using Scan to assign column data to struct fields
	for rows.Next() {
		var alb Album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
		}
		albums = append(albums, alb)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("albumsByArtist %q, %v", name, err)
	}

	return albums, nil
}

// albumByID queries for the album with the specified ID
func albumByID(id int64) (Album, error) {
	// An album to hold data from the returned row
	var alb Album

	row := db.QueryRow("SELECT * FROM album WHERE id = $1", id)
	if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
		if err == sql.ErrNoRows {
			return alb, fmt.Errorf("ablumById %d: no such album", id)
		}
		return alb, fmt.Errorf("ablumsById %d: %v", id, err)
	}
	return alb, nil
}

// addAlbum adds the specified album to the database
// returning the album ID of the new entry
func addAlbum(alb Album) (int64, error) {
	var albID int64
	row := db.QueryRow(
		"INSERT INTO album (title, artist, price) VALUES ($1, $2, $3) RETURNING id",
		alb.Title,
		alb.Artist,
		alb.Price,
	)
	if err := row.Scan(&albID); err != nil {
		return 0, fmt.Errorf("addAlbum: %v", err)
	}

	return albID, nil
}
