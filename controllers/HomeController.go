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

	// 如果有历史记录,将历史记录放在前面
	if len(historyBooks) > 0 {
		// 去重
		existMap := make(map[int]bool)
		for _, book := range historyBooks {
			existMap[book.BookId] = true
		}

		var filteredBooks []*models.BookResult
		for _, book := range books {
			if !existMap[book.BookId] {
				filteredBooks = append(filteredBooks, book)
			}
		}

		// 合并历史记录和其他书籍
		allBooks := append(historyBooks, filteredBooks...)

		// 根据分页截取需要的记录
		start := (pageIndex - 1) * pageSize
		end := start + pageSize
		if start < len(allBooks) {
			if end > len(allBooks) {
				end = len(allBooks)
			}
			books = allBooks[start:end]
		}
	}

	if totalCount > 0 {
		pager := pagination.NewPagination(c.Ctx.Request, totalCount, pageSize, c.BaseUrl())
		c.Data["PageHtml"] = pager.HtmlPages()
	} else {
		c.Data["PageHtml"] = ""
	}
	c.Data["TotalPages"] = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	c.Data["Lists"] = books
}
