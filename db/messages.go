package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/swap-messenger/Backend/models"
	// "github.com/AlexeyArno/Gologer"
)

func addMessage(userId int64, chatId int64, content string) (int64, error) {
	res, err := CheckUserInChatDelete(userId, chatId)
	if err != nil {
		return 0, errors.New("Check user in chat error: " + err.Error())
	}
	if res {
		return 0, errors.New("User deleted from chat")
	}
	m := Message{Author: &User{Id: userId}, Content: content, Chat: &Chat{Id: chatId}, Time: time.Now().Unix()}
	o.Begin()
	id, err := o.Insert(&m)
	if err != nil {
		o.Rollback()
		return 0, errors.New("Message insert error: " + err.Error())
	}
	o.Commit()
	return id, nil
}

func GetMessages(userId int64, chatId int64, add bool, lastIndex int64) ([]*models.NewMessageToUser, error) {
	type MessageTemplate struct {
		Id       int64
		Content  string
		AuthorId int64
		Name     string
		Login    string
		Time     int64
	}
	var templates []MessageTemplate
	var final []*models.NewMessageToUser
	const MAXTIME = 9999999999
	var cUser ChatUser
	err := o.QueryTable("chat_users").Filter("user_id", userId).
		Filter("chat_id", chatId).RelatedSel().One(&cUser)
	if err != nil {
		return final, errors.New("User is not in chat")
	}

	delTimes, err := cUser.GetDeletePoints()
	if err != nil {
		return final, errors.New("Can't decode delete points")
	}
	qb, _ := orm.NewQueryBuilder(driver)
	//Get message from db
	qb.Select("messages.id",
		"messages.content",
		"messages.author_id",
		"users.name",
		"users.login",
		"messages.time").
		From("messages").
		InnerJoin("users").On("messages.author_id = users.id").
		Where("messages.chat_id = ?")
	if cUser.Chat.Type != 2 {
		for i := 0; i < len(delTimes); i++ {
			if i == 0 && delTimes[0][0] == 0 {
				qb.And(fmt.Sprintf("((messages.time>=%d) and  (messages.time<=%d)) ", cUser.Start, MAXTIME))
			} else {
				if i == 0 {
					qb.And(fmt.Sprintf("((messages.time>=%d) and  (messages.time<=%d)) ", cUser.Start, delTimes[i][0]))
				} else if i > 0 {
					qb.And(fmt.Sprintf("((messages.time>=%d) and  (messages.time<=%d)) ", delTimes[i-1][1], delTimes[i][0]))
					if delTimes[i][0] == 0 {
						qb.And(fmt.Sprintf("((messages.time>=%d) and  (messages.time<=%d)) ", delTimes[i-1][1], MAXTIME))
					}
				}
			}
		}
		if add {
			qb.And(fmt.Sprintf("messages.id > %d", lastIndex))
		}
		qb.OrderBy("messages.time").Asc().Limit(80)
	}
	sql := qb.String()
	o.Raw(sql, chatId).QueryRows(&templates)

	//Get Content and File information
	for _, v := range templates {
		type ContentFirst struct {
			Message   string  `json:"content"`
			Documents []int64 `json:"documents"`
			Type      string  `json:"type"`
		}

		var Content ContentFirst
		err := json.Unmarshal([]byte(v.Content), &Content)
		if err != nil {
			// Gologer.PError("Fail unmarshal : "+v.Content)
			continue
		}
		var docs []map[string]interface{}
		for _, v := range Content.Documents {
			doc, err := GetFileInformation(v)
			if err != nil {
				continue
			}
			docs = append(docs, doc)
		}
		var mes models.MessageContentToUser

		mes.Documents = docs
		mes.Message = Content.Message
		mes.Type = Content.Type

		final = append(final, &models.NewMessageToUser{
			ID:          v.Id,
			ChatId:      chatId,
			AuthorId:    v.AuthorId,
			AuthorName:  v.Name,
			AuthorLogin: v.Login,
			Time:        v.Time,
			Content:     &mes})
	}
	return final, nil
}