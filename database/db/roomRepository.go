package db

import (
	"database/sql"

	"github.com/AkinAD/grpc_chat_fromNodeToGo/models/room"
)

type Room struct {
	Id      string
	Name    string
	Private bool
}

func (room *Room) GetId() string {
	return room.Id
}

func (room *Room) GetName() string {
	return room.Name
}

func (room *Room) GetPrivate() bool {
	return room.Private
}

type RoomRepository struct {
	Db *sql.DB
}

func (repo *RoomRepository) AddRoom(room room.Room) {
	stmt, err := repo.Db.Prepare("INSERT INTO room(id, name, private) values(?,?,?)")
	checkErr(err)

	_, err = stmt.Exec(room.GetId(), room.GetName(), room.GetPrivate())
	checkErr(err)
}

func (repo *RoomRepository) FindRoomByName(name string) *room.Room {

	row := repo.Db.QueryRow("SELECT id, name, private FROM room where name = ? LIMIT 1", name)

	var room room.Room

	if err := row.Scan(&room.ID, &room.Name, &room.Private); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}

	return &room

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
