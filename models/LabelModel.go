package models

import (
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/mindoc-org/mindoc/conf"
)

type Label struct {
	LabelId    int    `orm:"column(label_id);pk;auto;unique;description(项目标签id)" json:"label_id"`
	LabelName  string `orm:"column(label_name);size(50);unique;description(项目标签名称)" json:"label_name"`
	BookNumber int    `orm:"column(book_number);description(包涵项目数量)" json:"book_number"`
}

// TableName 获取对应数据库表名.
func (m *Label) TableName() string {
	return "label"
}

// TableEngine 获取数据使用的引擎.
func (m *Label) TableEngine() string {
	return "INNODB"
}

func (m *Label) TableNameWithPrefix() string {
	return conf.GetDatabasePrefix() + m.TableName()
}

func NewLabel() *Label {
	return &Label{}
}

func (m *Label) FindFirst(field string, value interface{}) (*Label, error) {
	o := orm.NewOrm()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter(field, value).One(m)

	return m, err
}

// 插入或更新标签.
func (m *Label) InsertOrUpdate(labelName string) error {
	o := orm.NewOrm()

	logs.Info("插入或更新标签: 查询标签是否存在 ->", labelName)

	// 先查询标签是否已存在
	err := o.QueryTable(m.TableNameWithPrefix()).Filter("label_name", labelName).One(m)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("查询标签时发生错误 ->", err)
		return err
	}

	if err == orm.ErrNoRows {
		// 如果标签不存在，创建一个新的标签
		logs.Info("标签不存在，准备创建新标签 ->", labelName)
		m.LabelName = labelName
		m.BookNumber = 0

		_, err = o.Insert(m)
		if err != nil {
			logs.Error("插入标签失败 ->", err)
			return err
		}
		logs.Info("成功插入标签 ->", labelName, "ID:", m.LabelId)
	} else {
		// 如果标签已存在，更新书籍数量
		logs.Info("标签已存在，ID:", m.LabelId, "名称:", m.LabelName)
		var count int
		err = o.Raw("SELECT COUNT(*) FROM md_book_label WHERE label_id = ?", m.LabelId).QueryRow(&count)
		if err == nil {
			m.BookNumber = count
			_, err = o.Update(m)
			logs.Info("更新标签 ->", labelName, "ID:", m.LabelId, "相关书籍数:", count)
			if err != nil {
				logs.Error("更新标签失败 ->", err)
				return err
			}
		}
	}
	return nil
}

// 批量插入或更新标签.
func (m *Label) InsertOrUpdateMulti(labels string) error {
	logs.Info("批量插入或更新标签 ->", labels)
	if labels != "" {
		labelArray := strings.Split(labels, ",")

		for _, label := range labelArray {
			if label != "" {
				logs.Info("插入或更新标签 ->", label)
				label = strings.TrimSpace(label)
				if label != "" {
					logs.Info("插入或更新标签 ->", label)
					if err := NewLabel().InsertOrUpdate(label); err != nil {
						logs.Error("批量插入或更新标签失败 ->", label, err)
						return err
					}
				}
			}
		}
	}
	return nil
}

// 删除标签
func (m *Label) Delete() error {
	o := orm.NewOrm()

	// 首先删除标签与书籍的关联关系
	if _, err := o.Raw("DELETE FROM "+NewBookLabel().TableNameWithPrefix()+" WHERE label_id= ?", m.LabelId).Exec(); err != nil {
		return err
	}

	// 然后删除标签本身
	_, err := o.Raw("DELETE FROM "+m.TableNameWithPrefix()+" WHERE label_id= ?", m.LabelId).Exec()
	if err != nil {
		return err
	}

	return nil
}

// 分页查找标签.
func (m *Label) FindToPager(pageIndex, pageSize int) (labels []*Label, totalCount int, err error) {
	o := orm.NewOrm()

	count, err := o.QueryTable(m.TableNameWithPrefix()).Count()

	if err != nil {
		return
	}
	totalCount = int(count)

	offset := (pageIndex - 1) * pageSize

	// 查询标签并同时更新书籍数量
	_, err = o.QueryTable(m.TableNameWithPrefix()).OrderBy("-book_number").Offset(offset).Limit(pageSize).All(&labels)

	if err == orm.ErrNoRows {
		logs.Info("没有查询到标签 ->", err)
		err = nil
		return
	}

	// 更新每个标签的书籍数量
	for i, label := range labels {
		var count int
		err = o.Raw("SELECT COUNT(*) FROM md_book_label WHERE label_id = ?", label.LabelId).QueryRow(&count)
		if err == nil {
			labels[i].BookNumber = count
			o.Update(label, "book_number")
		}
	}

	return
}

// FindByName 根据名称模糊搜索标签
func (m *Label) FindByName(name string) ([]*Label, error) {
	o := orm.NewOrm()

	var labels []*Label

	_, err := o.QueryTable(m.TableNameWithPrefix()).
		Filter("label_name__icontains", name).
		OrderBy("-book_number").
		Limit(10).All(&labels)

	return labels, err
}
