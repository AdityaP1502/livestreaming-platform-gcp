package public

import (
	"database/sql"

	"github.com/AdityaP1502/livestreaming-platform-gcp/api/go/base"
)

func insertNewUserToDatabase(user *Users, app *base.App) error {
	query := "INSERT INTO Users (FullName, Username, Password) VALUES (?, ?, ?)"
	_, err := app.Connection.Exec(query, user.FullName, user.Username, user.Password)

	return err
}

func usernameExists(username string, app *base.App) (bool, error) {
	// Query the database to check if the username exists
	query := "SELECT EXISTS(SELECT 1 FROM Users WHERE username = ?)"
	var exists bool

	err := app.Connection.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func retrieveUserInformation(user *Users, app *base.App) *sql.Row {
	query := "SELECT Password FROM Users WHERE Username = ?"
	return app.Connection.QueryRow(query, user.Username)
}

func retrieveUserID(user *Users, app *base.App) *sql.Row {
	query := "SELECT UserID FROM Users WHERE Username = ?"
	return app.Connection.QueryRow(query, user.Username)
}

func saveStreamData(stream *Stream, app *base.App) error {
	query := "INSERT INTO Streams (RTSP_URL, URL, ID, UserID, Status) VALUES (?, ?, ?, ?, ?)"
	_, err := app.Connection.Exec(query, stream.RTSPURL, stream.URL, stream.ID, stream.UserID, stream.Status)

	return err
}

func streamExist(username string, streamID string, app *base.App) *sql.Row {
	query := `SELECT s.StreamID
	FROM Streams s
	JOIN Users u ON s.UserID = u.UserID
	WHERE s.ID = ? AND u.Username = ?`

	return app.Connection.QueryRow(query, streamID, username)
}

func insertStreamMetadata(metadata *StreamMetadata, app *base.App) error {
	query := "INSERT INTO Streams_Metadata (Title, CreatedAt, Thumbnail, StreamID) VALUES (?, ?, ?, ?)"

	_, err := app.Connection.Exec(query, metadata.Title, metadata.CreatedAt, metadata.Thumbnail, metadata.StreamID)

	return err
}

func deleteStream(id int, app *base.App) error {
	query := `DELETE FROM Streams_Metadata WHERE StreamID =?`

	_, err := app.Connection.Exec(query, id)

	if err != nil {
		return err
	}

	query = `DELETE FROM Streams WHERE StreamID=?`
	_, err = app.Connection.Exec(query, id)

	return err
}

func updateStreamStatus(status string, id int, app *base.App) error {
	query := `UPDATE Streams
	SET Status=?
	WHERE StreamID =?`

	_, err := app.Connection.Exec(query, status, id)

	return err
}

func getAllStream(app *base.App) (*sql.Rows, error) {
	query := `SELECT Streams.URL,
    Users.Username,
    Streams_Metadata.Title,
	Streams_Metadata.CreatedAt,
	Streams_Metadata.Thumbnail
	FROM Streams
	JOIN Users ON Streams.UserID = Users.UserID
	JOIN Streams_Metadata ON Streams.StreamID = Streams_Metadata.StreamID
	WHERE Streams.Status =?`

	return app.Connection.Query(query, "active")
}
