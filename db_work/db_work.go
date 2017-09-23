package spatium_db_work

import (
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"crypto/sha256"
	"os"
	"fmt"
	models "github.com/AlexArno/spatium/models"
	"time"
	"errors"
	"encoding/json"
	"strconv"
)
var (
	activeConn *sql.DB
	activeConnIsReal bool
)


func GetInfo() string{
	return "Info"
}

func GetUser(s_type string, data map[string]string)(*models.User, error){
	user := new(models.User)
	if !activeConnIsReal{
		OpenDB()
	}
	if s_type == "login"{
		rows, err := activeConn.Prepare("SELECT id, login, pass, u_name FROM people WHERE (login=?) AND (pass=?)")
		if err != nil {
			panic(nil)
		}
		//make hash of user's password
		h := sha256.New()
		h.Write([]byte(data["pass"]))
		query := rows.QueryRow(data["login"], h.Sum(nil))

		err = query.Scan(&user.ID, &user.Login, &user.Pass, &user.Name)
		defer rows.Close()
		if err != nil {
			return nil, err
		}
		return user,nil
	}else{
		rows, err := activeConn.Prepare("SELECT id, login, u_name FROM people WHERE id=?")
		if err != nil {
			panic(nil)
		}
		query := rows.QueryRow(data["id"])
		fmt.Println(data["id"])
		err = query.Scan(&user.ID, &user.Login, &user.Name)
		if err == sql.ErrNoRows{
			return nil, err
		}
		defer rows.Close()
		if err != nil {
			return nil, err
		}
		return user,nil
	}

}

func CreateUser(login string, pass string, u_name string)(string, string, error){
	if !activeConnIsReal{
		OpenDB()
	}
	//test for equals logins
	var id_now string
	rows, err := activeConn.Prepare("SELECT id FROM people WHERE login=?")
	if err != nil {
		panic(nil)
	}
	query := rows.QueryRow(login).Scan(&id_now)
	defer rows.Close()
	if query != sql.ErrNoRows{
		return "","Login is busy",err
	}

	statement, err := activeConn.Prepare("INSERT INTO people (login, pass, u_name) VALUES (?, ?, ?)")
	if err != nil {
		return "","DB failed query",err
	}
	//make hash of user's password
	h := sha256.New()
	h.Write([]byte(pass))
	statement.Exec(login, h.Sum(nil), u_name)
	rows, err = activeConn.Prepare("SELECT id FROM people WHERE login=?")
	if err != nil {
		panic(nil)
	}
	query = rows.QueryRow(login).Scan(&id_now)
	if query == sql.ErrNoRows{
		return "","Some is fail",err
	}
	return id_now,"Success", nil
}

func InsertUserInChat(user_id string, chat_id int64)( error){
	if !activeConnIsReal{
		OpenDB()
	}
	var id_now string
	rows, err := activeConn.Prepare("SELECT chat_id FROM people_in_chats WHERE (user_id=?) AND (chat_id=?)")
	if err != nil {
		panic(nil)
	}
	query := rows.QueryRow(user_id, chat_id).Scan(&id_now)
	defer rows.Close()
	if query != sql.ErrNoRows{
		return errors.New("User already in chat")
	}
	statement, err := activeConn.Prepare("INSERT INTO people_in_chats (user_id, chat_id) VALUES (?, ?)")
	if err != nil {
		return errors.New("DB failed query")
	}
	//make hash of user's password
	statement.Exec(user_id, chat_id)
	statement, err = activeConn.Prepare("UPDATE chats SET lastmodify=? WHERE id=?")
	if err != nil {
		return errors.New("DB failed query")
	}
	//make hash of user's password
	statement.Exec(time.Now().Unix(), chat_id)
	return nil
}

func CreateChat(name string, author_id string)(string,  error){
	if !activeConnIsReal{
		OpenDB()
	}
	statement, err := activeConn.Prepare("INSERT INTO chats (name,  author_id,moders_ids, lastmodify) VALUES (?, ?, ?, ?)")
	if err != nil {
		return "",errors.New("Failed permanent statement")
	}
	//make hash of user's password
	res, err := statement.Exec(name,  author_id,"[]", time.Now().Unix())
	if err != nil {
		return "",errors.New("Failed exec statement")
	}
	id, _ := res.LastInsertId()
	err = InsertUserInChat(author_id, id)
	if err != nil {
		return "",err
		//fmt.Println(fin)
	}
	mess_mss := "Я создал этот чат"
	docs := []string{}
	m_type := "a_msg"
	mess := models.MessageContent{&mess_mss, &docs, &m_type}
	data ,err := json.Marshal(mess)
	if err != nil{
		return "", err
	}
	f_id,err := strconv.ParseFloat(author_id, 64)
	if err != nil{
		return "", err
	}
	err = AddMessage(f_id, float64(id), string(data))
	if err != nil{
		return "", err
	}
	return string(id), nil

}

func GetMyChats(user_id float64)([]*models.UserChatInfo, error){
	var chats_ids []*models.UserChatInfo
	var middle []map[string]string
	rows, err := activeConn.Query("SELECT chats.id, chats.name FROM people_in_chats INNER JOIN chats ON people_in_chats.chat_id = chats.id WHERE user_id=?", user_id)
	if err != nil {
		fmt.Println("Outside")
		return nil,err
	}
	defer rows.Close()
	for rows.Next(){
		var id, name string
		if err := rows.Scan(&id,  &name); err != nil {
			return nil,err
		}
		middle=append(middle, map[string]string{"id": id, "name": name})

	}
	for _,i := range middle{
		var author_name, content string
		message, err := activeConn.Prepare("SELECT  messages.content, people.u_name FROM messages INNER JOIN people ON messages.user_id = people.id WHERE chat_id=? ORDER BY time DESC")
		if err != nil {
			fmt.Println("Inside")
			return nil,err
		}
		query := message.QueryRow(i["id"])

		err = query.Scan(&content, &author_name)
		if err == sql.ErrNoRows{
			//return nil, err
			content = ""
			author_name = ""
		}
		f_id,err := strconv.ParseFloat(i["id"], 64)
		if err != nil {
			return nil, err
		}
		var m_content models.MessageContent
		err = json.Unmarshal([]byte(content), &m_content)
		if err!=nil{
			return nil, err
		}
		chats_ids=append(chats_ids, &models.UserChatInfo{f_id,i["name"], []string{}, author_name, &m_content,0 })
		defer message.Close()
		//chats_ids
	}
	if err := rows.Err(); err != nil {
		return nil,err
	}
	return chats_ids, nil
}

func AddMessage(user_id float64, chat_id float64, content string)(error){
	if !activeConnIsReal{
		OpenDB()
	}
	// Is user in chat?
	err := CheckUserINChat(user_id, chat_id)
	if err != nil{
		return err
	}
//	Create message
	statement, err := activeConn.Prepare("INSERT INTO messages (user_id, chat_id, content, time) VALUES (?, ?, ?, ?)")
	if err != nil {
		return errors.New("DB failed query")
	}
	//make hash of user's password
	_, err = statement.Exec(user_id, chat_id, content, time.Now().Unix())
	if err != nil {
		return errors.New("Failed exec statement")
	}
	return nil
}

func CheckUserINChat(user_id float64, chat_id float64)(error){
	var id_now string
	rows, err := activeConn.Prepare("SELECT chat_id FROM people_in_chats WHERE (user_id=?) AND (chat_id=?)")
	if err != nil {
		panic(nil)
	}
	query := rows.QueryRow(user_id, chat_id).Scan(&id_now)
	defer rows.Close()
	if query == sql.ErrNoRows{
		return errors.New("User aren't in chat")
	}
	return nil
}

func GetMessages(chat_id float64)([]models.NewMessageToUser, error){
	var messages []models.NewMessageToUser
	rows, err := activeConn.Query("SELECT messages.user_id, messages.content, messages.chat_id,   people.u_name  FROM messages INNER JOIN people ON messages.user_id = people.id WHERE messages.chat_id=?", chat_id)
	if err != nil {
		return nil,err
	}
	defer rows.Close()
	for rows.Next(){
		var id, content, u_name, c_id string
		if err := rows.Scan(&id,  &content,&c_id, &u_name); err != nil {
			return nil,err
		}
		//decode content
		var r_content *models.MessageContent
		err = json.Unmarshal([]byte(content), &r_content)
		if err != nil{
			return nil,err
		}
		f64_c_id, err := strconv.ParseFloat(c_id, 64)
		if err != nil {
			return nil,err
		}
		f64_id, err := strconv.ParseFloat(id, 64)
		if err != nil {
			return nil,err
		}
		messages = append(messages, models.NewMessageToUser{&f64_c_id,r_content,&f64_id,&u_name})
	}
	return messages, nil
}

func createDB_structs(database *sql.DB) {
	//Create user structs
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS people (id INTEGER PRIMARY KEY, login TEXT, pass TEXT, u_name TEXT)")
	statement.Exec()
	user_id, fin, err := CreateUser("god", "1111", "Alex")
	if err != nil {
		fmt.Println(fin)
		return
	}
	//Create people in chat structs

	statement, _ = database.Prepare("CREATE TABLE IF NOT EXISTS people_in_chats ( user_id INTEGER, chat_id INTEGER)")
	statement.Exec()

	//Create messages structs
	statement, _ = database.Prepare("CREATE TABLE IF NOT EXISTS messages (id INTEGER PRIMARY KEY, user_id INTEGER, chat_id INTEGER, content TEXT, time INTEGER)")
	statement.Exec()

	//Create chat structs
	statement, _ = database.Prepare("CREATE TABLE IF NOT EXISTS chats (id INTEGER PRIMARY KEY, name TEXT,  author_id INTEGER , moders_ids TEXT, lastmodify INTEGER)")
	statement.Exec()
	_, err = CreateChat("globalChat",  user_id)
	if err != nil {
		fmt.Println(err.Error())
	}



	}


func OpenDB(){
	newDB := false
	_, err := os.Open("app.db")
	if err != nil{
		newDB = true
		file, err := os.Create("app.db")
		if err != nil {
			// handle the error here
			fmt.Println("God: i cant create database, your PC is atheist")
			return
		}
		defer file.Close()
		fmt.Println("God: im create database")
	}
	database, _ := sql.Open("sqlite3", "./app.db")
	if newDB{
		createDB_structs(database)
	}
	activeConn = database
	activeConnIsReal=true
}

