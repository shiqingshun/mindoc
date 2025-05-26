package controllers

import (
	"math"
	"net/url"
	"sort"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mindoc-org/mindoc/conf"
	"github.com/mindoc-org/mindoc/models"
	"github.com/mindoc-org/mindoc/utils/pagination"
)

type HomeController struct {
	BaseController
}

func (c *HomeController) Prepare() {
	c.BaseController.Prepare()
	//如果没有开启匿名访问，则跳转到登录页面
	if !c.EnableAnonymous && c.Member == nil {
		c.Redirect(conf.URLFor("AccountController.Login")+"?url="+url.PathEscape(conf.BaseUrl+c.Ctx.Request.URL.RequestURI()), 302)
	}
}

func (c *HomeController) Index() {
	c.Prepare()
	c.TplName = "home/index.tpl"

	pageIndex, _ := c.GetInt("page", 1)
	pageSize := 18
	memberId := 0
	if c.Member != nil {
		memberId = c.Member.MemberId
	}
	// 获取用户的浏览历史
	var historyBooks []*models.BookResult
	if memberId > 0 {
		histories, _, err := models.NewBookReadHistory().FindToPager(memberId, 1, 100)
		if err == nil && len(histories) > 0 {
			// 构建书籍ID到最后阅读时间的映射
			bookTimeMap := make(map[int]time.Time)
			var bookIds []int
			for _, history := range histories {
				bookIds = append(bookIds, history.BookId)
				bookTimeMap[history.BookId] = history.LastReadTime
			}

			// 根据浏览历史获取书籍详情
			if len(bookIds) > 0 {
				books, _ := models.NewBook().GetBooksByIds(bookIds)
				if len(books) > 0 {
					for _, book := range books {
						result := models.NewBookResult().ToBookResult(*book)
						historyBooks = append(historyBooks, result)
					}
					// 按照最后阅读时间排序
					sort.Slice(historyBooks, func(i, j int) bool {
						return bookTimeMap[historyBooks[i].BookId].After(bookTimeMap[historyBooks[j].BookId])
					})
				}
			}
		}
	}

	// 获取其他书籍
	books, totalCount, err := models.NewBook().FindForHomeToPager(pageIndex, pageSize, memberId)
	if err != nil {
		logs.Error(err)
		c.Abort("500")
	}

	// 获取所有项目空间
	itemsets, _, err := models.NewItemsets().FindToPager(1, 1000)
	if err != nil {
		logs.Error("获取项目空间失败:", err)
	}

	// 创建项目空间映射
	itemsetsMap := make(map[int]*models.Itemsets)
	for _, item := range itemsets {
		itemsetsMap[item.ItemId] = item
	}

	// 按项目空间分组书籍
	groupedBooks := make(map[int][]*models.BookResult)
	groupedOrder := make([]int, 0)
	groupedBooks[0] = make([]*models.BookResult, 0) // 未分组

	// 统计所有书籍分组
	for _, book := range books {
		if book.ItemId > 0 {
			if _, ok := groupedBooks[book.ItemId]; !ok {
				groupedBooks[book.ItemId] = make([]*models.BookResult, 0)
			}
			groupedBooks[book.ItemId] = append(groupedBooks[book.ItemId], book)
		} else {
			groupedBooks[0] = append(groupedBooks[0], book)
		}
	}

	// 统计历史涉及的 itemId，按历史顺序去重
	historyItemIdSet := make(map[int]bool)
	historyItemOrder := make([]int, 0)
	for _, book := range historyBooks {
		if book.ItemId > 0 && !historyItemIdSet[book.ItemId] {
			historyItemIdSet[book.ItemId] = true
			historyItemOrder = append(historyItemOrder, book.ItemId)
		}
	}

	// 先按历史顺序排列分组
	for _, itemId := range historyItemOrder {
		if _, ok := groupedBooks[itemId]; ok {
			groupedOrder = append(groupedOrder, itemId)
		}
	}
	// 再补充未在历史中的分组
	for itemId := range groupedBooks {
		if itemId == 0 {
			continue
		}
		if !historyItemIdSet[itemId] {
			groupedOrder = append(groupedOrder, itemId)
		}
	}
	// 最后未分组的（itemId==0）
	if len(groupedBooks[0]) > 0 {
		groupedOrder = append(groupedOrder, 0)
	}

	if totalCount > 0 {
		pager := pagination.NewPagination(c.Ctx.Request, totalCount, pageSize, c.BaseUrl())
		c.Data["PageHtml"] = pager.HtmlPages()
	} else {
		c.Data["PageHtml"] = ""
	}
	c.Data["TotalPages"] = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	c.Data["Lists"] = books
	c.Data["GroupedBooks"] = groupedBooks
	c.Data["ItemsetsMap"] = itemsetsMap
	c.Data["GroupedOrder"] = groupedOrder
}
