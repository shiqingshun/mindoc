package models

import (
	"context"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/mindoc-org/mindoc/conf"
)

// BookItemRelationship 书籍和项目空间的多对多关联表
type BookItemRelationship struct {
	RelationshipId int       `orm:"pk;auto;unique;column(relationship_id)" json:"relationship_id"`
	BookId         int       `orm:"column(book_id);type(int);index;description(书籍ID)" json:"book_id"`
	ItemId         int       `orm:"column(item_id);type(int);index;description(项目空间ID)" json:"item_id"`
	CreateTime     time.Time `orm:"column(create_time);type(datetime);auto_now_add;description(创建时间)" json:"create_time"`
}

// TableName 获取对应数据库表名
func (m *BookItemRelationship) TableName() string {
	return "book_item_relationship"
}

// TableEngine 获取数据使用的引擎
func (m *BookItemRelationship) TableEngine() string {
	return "INNODB"
}

// TableNameWithPrefix 获取带前缀的表名
func (m *BookItemRelationship) TableNameWithPrefix() string {
	return conf.GetDatabasePrefix() + m.TableName()
}

// NewBookItemRelationship 创建一个新的书籍与项目空间关系对象
func NewBookItemRelationship() *BookItemRelationship {
	return &BookItemRelationship{}
}

// Insert 添加一个书籍与项目空间的关系
func (m *BookItemRelationship) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(m)
	return err
}

// Delete 删除书籍与项目空间的关系
func (m *BookItemRelationship) Delete() error {
	o := orm.NewOrm()
	_, err := o.Delete(m)
	return err
}

// GetItemIdsByBookId 根据书籍ID获取所有关联的项目空间ID
func (m *BookItemRelationship) GetItemIdsByBookId(bookId int) ([]int, error) {
	o := orm.NewOrm()
	var itemIds []int
	_, err := o.Raw("SELECT item_id FROM "+m.TableNameWithPrefix()+" WHERE book_id = ?", bookId).QueryRows(&itemIds)
	return itemIds, err
}

// GetBookIdsByItemId 根据项目空间ID获取所有关联的书籍ID
func (m *BookItemRelationship) GetBookIdsByItemId(itemId int) ([]int, error) {
	o := orm.NewOrm()
	var bookIds []int
	_, err := o.Raw("SELECT book_id FROM "+m.TableNameWithPrefix()+" WHERE item_id = ?", itemId).QueryRows(&bookIds)
	return bookIds, err
}

// DeleteByBookId 删除指定书籍的所有项目空间关系
func (m *BookItemRelationship) DeleteByBookId(bookId int) error {
	o := orm.NewOrm()
	_, err := o.Raw("DELETE FROM "+m.TableNameWithPrefix()+" WHERE book_id = ?", bookId).Exec()
	return err
}

// UpdateByBookId 更新指定书籍的项目空间关系
func (m *BookItemRelationship) UpdateByBookId(bookId int, itemIds []int) error {
	o := orm.NewOrm()

	// 开启事务
	err := o.DoTx(func(ctx context.Context, txOrm orm.TxOrmer) error {
		// 先删除原有关系
		if _, err := txOrm.Raw("DELETE FROM "+m.TableNameWithPrefix()+" WHERE book_id = ?", bookId).Exec(); err != nil {
			return err
		}

		// 添加新关系
		for _, itemId := range itemIds {
			if _, err := txOrm.Raw("INSERT INTO "+m.TableNameWithPrefix()+" (book_id, item_id, create_time) VALUES (?, ?, ?)",
				bookId, itemId, time.Now()).Exec(); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		logs.Error("更新书籍项目空间关系失败 -> bookId=", bookId, err)
	}
	return err
}

// GetItemNamesByBookId 根据书籍ID获取所有关联的项目空间名称
func (m *BookItemRelationship) GetItemNamesByBookId(bookId int) (map[int]string, error) {
	o := orm.NewOrm()
	var results []struct {
		ItemId   int    `orm:"column(item_id)"`
		ItemName string `orm:"column(item_name)"`
	}

	sql := `SELECT bir.item_id, its.item_name 
			FROM ` + m.TableNameWithPrefix() + ` AS bir 
			LEFT JOIN ` + NewItemsets().TableNameWithPrefix() + ` AS its ON bir.item_id = its.item_id 
			WHERE bir.book_id = ?`

	_, err := o.Raw(sql, bookId).QueryRows(&results)
	if err != nil {
		return nil, err
	}

	itemNames := make(map[int]string)
	for _, item := range results {
		itemNames[item.ItemId] = item.ItemName
	}

	return itemNames, nil
}
