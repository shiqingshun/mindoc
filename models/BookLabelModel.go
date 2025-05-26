package models

import (
	"errors"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/mindoc-org/mindoc/conf"
)

// BookLabel 项目标签对应关系表
type BookLabel struct {
	BookLabelId int `orm:"column(book_label_id);pk;auto;unique;description(自增ID)" json:"book_label_id"`
	BookId      int `orm:"column(book_id);type(int);description(项目ID)" json:"book_id"`
	LabelId     int `orm:"column(label_id);type(int);description(标签ID)" json:"label_id"`
}

// TableName 获取对应数据库表名
func (m *BookLabel) TableName() string {
	return "book_label"
}

// TableEngine 获取数据使用的引擎
func (m *BookLabel) TableEngine() string {
	return "INNODB"
}

func (m *BookLabel) TableNameWithPrefix() string {
	return conf.GetDatabasePrefix() + m.TableName()
}

// TableUnique 联合唯一键
func (m *BookLabel) TableUnique() [][]string {
	return [][]string{
		{"book_id", "label_id"},
	}
}

func (m *BookLabel) QueryTable() orm.QuerySeter {
	return orm.NewOrm().QueryTable(m.TableNameWithPrefix())
}

// NewBookLabel 新建项目标签关联
func NewBookLabel() *BookLabel {
	return &BookLabel{}
}

// GetBookLabels 根据项目ID获取所有标签
func (m *BookLabel) GetBookLabels(bookId int) ([]*Label, error) {
	o := orm.NewOrm()

	var labels []*Label

	// 查询该书籍的所有标签ID
	sql := `SELECT label.* FROM ` + m.TableNameWithPrefix() + ` AS rel 
			LEFT JOIN ` + NewLabel().TableNameWithPrefix() + ` AS label ON rel.label_id = label.label_id
			WHERE rel.book_id = ? ORDER BY label.label_name`

	_, err := o.Raw(sql, bookId).QueryRows(&labels)

	if err != nil && err != orm.ErrNoRows {
		logs.Error("获取书籍标签时出错 ->", err)
		return nil, err
	}

	return labels, nil
}

// InsertOrUpdateMulti 批量插入或更新标签关联
func (m *BookLabel) InsertOrUpdateMulti(bookId int, labelNames string) error {
	if bookId <= 0 {
		return ErrInvalidParameter
	}

	o := orm.NewOrm()

	// 首先删除当前书籍所有标签关联
	_, err := o.QueryTable(m.TableNameWithPrefix()).Filter("book_id", bookId).Delete()

	if err != nil {
		return err
	}

	// 如果没有标签，则直接返回
	if strings.TrimSpace(labelNames) == "" {
		return nil
	}

	labelNameList := strings.Split(labelNames, ",")

	// 创建或获取每个标签，并创建关联
	for _, labelName := range labelNameList {
		labelName = strings.TrimSpace(labelName)
		if labelName == "" {
			continue
		}

		// 插入或获取标签
		label := NewLabel()
		err = label.InsertOrUpdate(labelName)
		if err != nil {
			return err
		}

		// 创建关联
		bookLabel := NewBookLabel()
		bookLabel.BookId = bookId
		bookLabel.LabelId = label.LabelId

		// 插入关联
		_, err = o.Insert(bookLabel)
		if err != nil {
			// 可能是因为唯一键约束导致的插入失败，忽略这种错误
			if strings.Contains(err.Error(), "Duplicate") {
				continue
			}
			return err
		}
	}

	return nil
}

// GetLabelBooks 获取指定标签的书籍列表
func (m *BookLabel) GetLabelBooks(labelId int, pageIndex, pageSize, memberId int) ([]*BookResult, int, error) {
	o := orm.NewOrm()

	var books []*BookResult
	var count int

	offset := (pageIndex - 1) * pageSize

	// 如果是登录用户
	if memberId > 0 {
		sql1 := `SELECT COUNT(book.book_id)
FROM md_books AS book
  LEFT JOIN md_book_label AS bl ON bl.book_id = book.book_id
  LEFT JOIN md_relationship AS rel ON rel.book_id = book.book_id AND rel.member_id = ?
  left join (select book_id,min(role_id) AS role_id
             from (select book_id,role_id
                   from md_team_relationship as mtr
                     left join md_team_member as mtm on mtm.team_id=mtr.team_id and mtm.member_id=? order by role_id desc )
as t group by book_id) as team on team.book_id=book.book_id
WHERE (book.privately_owned = 0 OR rel.role_id >=0 OR team.role_id >=0) 
AND bl.label_id = ?`

		err := o.Raw(sql1, memberId, memberId, labelId).QueryRow(&count)
		if err != nil {
			return nil, 0, err
		}

		sql2 := `SELECT book.*,rel1.*,mdmb.account AS create_name,mdmb.real_name FROM md_books AS book
  LEFT JOIN md_book_label AS bl ON bl.book_id = book.book_id
  LEFT JOIN md_relationship AS rel ON rel.book_id = book.book_id AND rel.member_id = ?
  left join (select book_id,min(role_id) AS role_id
             from (select book_id,role_id
                   from md_team_relationship as mtr
                     left join md_team_member as mtm on mtm.team_id=mtr.team_id and mtm.member_id=? order by role_id desc )
as t group by book_id) as team on team.book_id=book.book_id
  LEFT JOIN md_relationship AS rel1 ON rel1.book_id = book.book_id AND rel1.role_id = 0
  LEFT JOIN md_members AS mdmb ON rel1.member_id = mdmb.member_id
WHERE (book.privately_owned = 0 OR rel.role_id >=0 OR team.role_id >=0) 
AND bl.label_id = ? ORDER BY order_index DESC ,book.book_id DESC limit ? offset ?`

		_, err = o.Raw(sql2, memberId, memberId, labelId, pageSize, offset).QueryRows(&books)
		if err != nil {
			return nil, 0, err
		}

	} else {
		// 如果是未登录用户，只查询公开的书籍
		countSql := `SELECT COUNT(book.book_id) FROM md_books AS book 
					LEFT JOIN md_book_label AS bl ON bl.book_id = book.book_id 
					WHERE book.privately_owned = 0 AND bl.label_id = ?`

		err := o.Raw(countSql, labelId).QueryRow(&count)
		if err != nil {
			return nil, 0, err
		}

		sql := `SELECT book.*,rel.*,mdmb.account AS create_name,mdmb.real_name FROM md_books AS book
				LEFT JOIN md_book_label AS bl ON bl.book_id = book.book_id 
				LEFT JOIN md_relationship AS rel ON rel.book_id = book.book_id AND rel.role_id = 0
				LEFT JOIN md_members AS mdmb ON rel.member_id = mdmb.member_id
				WHERE book.privately_owned = 0 AND bl.label_id = ? 
				ORDER BY order_index DESC ,book.book_id DESC limit ? offset ?`

		_, err = o.Raw(sql, labelId, pageSize, offset).QueryRows(&books)
		if err != nil {
			return nil, 0, err
		}
	}

	return books, count, nil
}

// GetBookLabelIds 获取书籍的所有标签ID
func (m *BookLabel) GetBookLabelIds(bookId int) ([]int, error) {
	o := orm.NewOrm()

	var labelIds []int

	_, err := o.Raw("SELECT label_id FROM "+m.TableNameWithPrefix()+" WHERE book_id = ?", bookId).QueryRows(&labelIds)

	if err != nil && err != orm.ErrNoRows {
		return nil, err
	}

	return labelIds, nil
}

// DeleteByBookId 删除书籍的所有标签关系
func (m *BookLabel) DeleteByBookId(bookId int) error {
	o := orm.NewOrm()

	_, err := o.QueryTable(m.TableNameWithPrefix()).Filter("book_id", bookId).Delete()

	return err
}

// GetLabelNamesByBookId 获取书籍的标签名称列表，以逗号分隔
func (m *BookLabel) GetLabelNamesByBookId(bookId int) string {
	labels, err := m.GetBookLabels(bookId)
	if err != nil {
		return ""
	}

	var labelNames []string

	for _, label := range labels {
		labelNames = append(labelNames, label.LabelName)
	}

	return strings.Join(labelNames, ",")
}

// UpdateBookLabels 更新书籍的标签关系
func (m *BookLabel) UpdateBookLabels(bookId int, labelIds []int) error {
	if bookId <= 0 {
		logs.Error("无效的书籍ID: ", bookId)
		return ErrInvalidParameter
	}

	o := orm.NewOrm()
	logs.Info("开始更新书籍标签关系，书籍ID: ", bookId, "标签数量: ", len(labelIds))

	// 首先删除当前书籍所有标签关联
	res, err := o.QueryTable(m.TableNameWithPrefix()).Filter("book_id", bookId).Delete()
	if err != nil {
		logs.Error("删除书籍原有标签关系失败: ", err)
		return err
	}
	logs.Info("已删除书籍原有标签关系: ", res, "条记录")

	// 如果没有新的标签，则直接返回
	if len(labelIds) == 0 {
		logs.Info("没有新的标签关系需要创建")
		return nil
	}
	// 批量插入新的标签关联
	var bookLabels []BookLabel

	// 验证每个标签是否存在
	for _, labelId := range labelIds {
		if labelId <= 0 {
			logs.Warn("忽略无效的标签ID: ", labelId)
			continue
		}

		// 验证标签是否存在
		label := NewLabel()
		err := o.QueryTable(label.TableNameWithPrefix()).Filter("label_id", labelId).One(label)
		if err != nil {
			if err == orm.ErrNoRows {
				logs.Warn("标签不存在，ID: ", labelId)
				continue
			}
			logs.Error("查询标签失败: ", err, "标签ID: ", labelId)
			continue
		}

		bookLabel := BookLabel{
			BookId:  bookId,
			LabelId: labelId,
		}
		bookLabels = append(bookLabels, bookLabel)
		logs.Info("准备添加书籍标签关系: 书籍ID=", bookId, "标签ID=", labelId, "标签名=", label.LabelName)
	}

	if len(bookLabels) > 0 {
		// 使用事务进行批量插入，确保数据一致性
		tx, err := o.Begin()
		if err != nil {
			logs.Error("开启事务失败: ", err)
			return err
		}

		success := true
		for _, bookLabel := range bookLabels {
			bl := BookLabel{
				BookId:  bookLabel.BookId,
				LabelId: bookLabel.LabelId,
			}
			_, err := tx.Insert(&bl)
			if err != nil {
				logs.Error("事务中插入书籍标签关系失败: ", err, "书籍ID=", bl.BookId, "标签ID=", bl.LabelId)
				success = false
				break
			}
			logs.Info("事务中插入书籍标签关系: 书籍ID=", bl.BookId, "标签ID=", bl.LabelId)
		}

		if success {
			if err := tx.Commit(); err != nil {
				logs.Error("提交事务失败: ", err)
				tx.Rollback()
				return err
			}
			logs.Info("批量插入书籍标签关系成功: ", len(bookLabels), "条记录")
		} else {
			tx.Rollback()
			logs.Error("回滚事务: 批量插入书籍标签关系失败")
			return errors.New("批量插入书籍标签关系失败")
		}
	}

	return nil
}
