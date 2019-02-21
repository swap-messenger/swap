package db

import (
	"errors"
	"strconv"
	"strings"

	"github.com/astaxie/beego/orm"
)

func GetUsersForCreateDialog(userId int64, name string) ([]map[string]interface{}, error) {
	//user which have dialogs with our
	var final []map[string]interface{}
	var userBuf []User
	var occupyUsers []int64
	var userDialogs []int64
	var occupyUsersStrings []string
	var userDialogsStrings []string
	qb, _ := orm.NewQueryBuilder(driver)
	//Delete users in caht
	qb.Select("chat_users.chat_id").
		From("chat_users").
		InnerJoin("chats").On("chats.id = chat_users.chat_id").
		Where("chats.type = 1")
	sql := qb.String()
	o.Raw(sql).QueryRows(&userDialogs)

	qb, _ = orm.NewQueryBuilder(driver)

	for _, v := range userDialogs {
		userDialogsStrings = append(userDialogsStrings, strconv.FormatInt(v, 10))
	}
	//Get users id  in users's dialogs
	s1 := strings.Join(userDialogsStrings, ",")
	qb.Select("chat_users.user_id").
		From("chat_users").
		Where("chat_users.chat_id").In(s1).
		And("chat_users.list__invisible = 0").
		And("chat_users.user_id = ?")
	sql = qb.String()
	o.Raw(sql, userId).QueryRows(&occupyUsers)
	for _, v := range occupyUsers {
		occupyUsersStrings = append(occupyUsersStrings, strconv.FormatInt(v, 10))
	}
	//Get users id  in users's dialogs
	s1 = strings.Join(occupyUsersStrings, ",")
	qb, _ = orm.NewQueryBuilder(driver)
	qb.Select("id", "name", "login").
		From("users").
		Where("id not").In(s1).
		And("chat_users.user_id <> ?").
		And("name LIKE ?").
		Or("login LIKE ?")
	sql = qb.String()
	o.Raw(sql, userId, name, name).QueryRows(&userBuf)

	for _, v := range userBuf {
		final = append(final, map[string]interface{}{
			"id": v.Id, "name": v.Name, "login": v.Login})
	}
	return final, nil
}

func HaveAlreadyDialog(userId int64, anotherUserId int64) (int64, error) {
	var final int64
	qb, _ := orm.NewQueryBuilder(driver)
	//Delete users in caht
	qb.Select("chat_id").
		From("dialogs").
		Where("user1 = ? and user2=?").
		Or("user2 = ? and user1=?")
	sql := qb.String()
	o.Raw(sql, userId, anotherUserId, userId, anotherUserId).QueryRow(&final)
	return final, nil
}

func CreateDialog(userId int64, anotherUserId int64) error {
	res, err := HaveAlreadyDialog(userId, anotherUserId)
	if err != nil {
		return err
	}
	if res != 0 {
		return errors.New("dialog already create")
	}
	cId, err := CreateChat("", userId)
	if err != nil {
		return err
	}
	err = InsertUserInChat(userId, cId, false)
	if err != nil {
		return err
	}
	err = InsertUserInChat(anotherUserId, cId, false)
	if err != nil {
		return err
	}
	d := Dialog{ChatId: cId, User1: &User{Id: userId}, User2: &User{Id: anotherUserId}}
	_, err = o.Insert(&d)
	if err != nil {
		return err
	}
	return nil
}