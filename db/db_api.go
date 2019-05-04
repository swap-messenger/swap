package db

import (
	"github.com/jinzhu/gorm"
	// _ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/swap-messenger/swap/settings"
)

type userRequestedCallback = func(userID int64, chatID int64, messageCommand int)
type chatCreatedCallback = func(AuthorId int64)

var (
	db                  gorm.DB
	UserRequestedToChat userRequestedCallback = nil
	ChatCreated         chatCreatedCallback   = nil
)

func LoadDb() {
	// register model

}

func createDB() error {
	// err := orm.RunSyncdb("default", true, false)
	// if err != nil {
	// 	return err
	// }
	// o = orm.NewOrm()
	// var sys System
	// sys.Date = time.Now().Unix()
	// sys.Version = "0.0.1"
	// _, err = o.Insert(&sys)
	// if err == nil {
	// 	return err
	// }
	return nil
}

func BeginDB() error {

	// orm.RegisterDriver("sqlite3", orm.DRSqlite)
	sett, err := settings.GetSettings()
	if err != nil {
		panic(err)
	}
	if sett.Backend.Test {
		// orm.Debug = true
		// orm.RegisterDataBase("default", "sqlite3", "file:test.db")
		db, err := gorm.Open("sqlite3", "test.db")
	} else {
		db, err := gorm.Open("sqlite3", settings.ServiceSettings.DB.SQLite.Path)
		//orm.RegisterDataBase("default", "sqlite3", "file:"+settings.ServiceSettings.DB.SQLite.Path)
	}
	if err != nil {
		panic("Failed connect")
	}

	db.AutoMigrate(&User{})
	db.AutoMigrate(&Chat{})
	db.AutoMigrate(&ChatUser{})
	db.AutoMigrate(&Message{})
	db.AutoMigrate(&File{})
	db.AutoMigrate(&System{})
	db.AutoMigrate(&Dialog{})

	// orm.RegisterModel(new(User))
	// orm.RegisterModel(new(Chat))
	// orm.RegisterModel(new(ChatUser))
	// orm.RegisterModel(new(Message))
	// orm.RegisterModel(new(File))
	// orm.RegisterModel(new(System))
	// orm.RegisterModel(new(Dialog))

	// o = orm.NewOrm()
	// sys := System{}
	// err = o.QueryTable("sys").Filter("id", 1).One(&sys)
	// if err != nil {
	// 	err = createDB()
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}
