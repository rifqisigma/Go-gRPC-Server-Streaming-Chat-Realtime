package main

import (
	"chat_api/cmd/database"
	"chat_api/entity"
	"log"
)

func main() {

	database.ConnectDB()

	if database.DB == nil {
		log.Fatal("❌ Database belum diinisialisasi")
	}

	err := database.DB.AutoMigrate(&entity.User{}, entity.ChatGroup{}, entity.GroupMember{}, entity.Chat{}, entity.ChatRead{})
	if err != nil {
		log.Fatalf("gagal migrasi boy %v", err)
	}

	log.Println("berhasil migrasi")

}
