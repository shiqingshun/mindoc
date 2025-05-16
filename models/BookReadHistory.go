package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// 用户阅读历史
type BookReadHistory struct {
	HistoryId    int       `orm:"column(history_id);pk;auto;unique;description(唯一标识符)" json:"history_id"`
	BookId       int       `orm:"column(book_id);type(int);description(书籍ID)" json:"book_id"`
	MemberId     int       `orm:"column(member_id);type(int);description(会员ID)" json:"member_id"`
	CreateTime   time.Time `orm:"column(create_time);type(datetime);auto_now_add;description(创建时间)" json:"create_time"`
	LastReadTime time.Time `orm:"column(last_read_time);type(datetime);auto_now;description(最近阅读时间)" json:"last_read_time"`
	ReadCount    int       `orm:"column(read_count);type(int);description(阅读次数)" json:"read_count"`
}

func NewBookReadHistory() *BookReadHistory {
	return &BookReadHistory{}
}

func (m *BookReadHistory) TableName() string {
	return "book_read_history"
}

func (m *BookReadHistory) TableNameWithPrefix() string {
	return "md_" + m.TableName()
}

func (m *BookReadHistory) Init() error {
	o := orm.NewOrm()
	return o.QueryTable(m.TableNameWithPrefix()).Filter("history_id", m.HistoryId).One(m)
}

func (m *BookReadHistory) Update(cols ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(m, cols...)
	return err
}

func (m *BookReadHistory) Add() error {
	o := orm.NewOrm()
	_, err := o.Insert(m)
	return err
}

// 获取或创建阅读历史
func (m *BookReadHistory) GetOrCreate(memberId, bookId int) (*BookReadHistory, error) {
	o := orm.NewOrm()

	history := &BookReadHistory{}
	err := o.QueryTable(m.TableNameWithPrefix()).Filter("member_id", memberId).Filter("book_id", bookId).One(history)

	if err == orm.ErrNoRows {
		history.BookId = bookId
		history.MemberId = memberId
		history.ReadCount = 1
		err = history.Add()
		return history, err
	}

	if err != nil {
		return history, err
	}

	history.ReadCount = history.ReadCount + 1
	history.LastReadTime = time.Now()
	err = history.Update()

	return history, err
}

// 获取用户的所有阅读历史
func (m *BookReadHistory) FindToPager(memberId, pageIndex, pageSize int) ([]*BookReadHistory, int, error) {
	o := orm.NewOrm()

	var histories []*BookReadHistory

	offset := (pageIndex - 1) * pageSize

	totalCount, err := o.QueryTable(m.TableNameWithPrefix()).Filter("member_id", memberId).Count()

	if err != nil {
		return histories, 0, err
	}

	_, err = o.QueryTable(m.TableNameWithPrefix()).Filter("member_id", memberId).
		OrderBy("-last_read_time").Offset(offset).Limit(pageSize).All(&histories)

	return histories, int(totalCount), err
}
