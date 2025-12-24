package controller

import (
	"MuXi/Library/config"
	"MuXi/Library/models"
	"MuXi/Library/utils"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary		用户注册
// @Description	注册用户
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		models.RegisterRequest	true	"注册请求"
// @Success		200		{object}	models.Response			"注册成功"
// @Failure		400		{object}	models.Response			"请求参数错误"
// @Failure		409		{object}	models.Response			"用户已存在"
// @Failure		500		{object}	models.Response			"服务器错误"
// @Router			/api/auth/register [post]
func Register(c *gin.Context) {
	var req models.RegisterRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    http.StatusBadRequest, //400
			Message: "请求参数错误",
		})
		return
	}

	var searchedUser models.User
	err = config.DB.Where("username = ?", req.Username).First(&searchedUser).Error
	//这里不用first的话就要用users切片，然后Find(&users)
	//我们只需要自己确保只有一个就ok
	if err == nil {
		c.JSON(http.StatusConflict, models.Response{
			Code:    http.StatusConflict, //409
			Message: "用户已存在",
		})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "查询数据库失败：" + err.Error(),
		})
		return
	}
	//notfound就可以注册了

	hashedPsw, err := utils.Hashtool(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError,
			Message: "注册密码哈希失败：" + err.Error(),
		})
		return
	}

	newUser := models.User{
		Username: req.Username,
		Password: hashedPsw,
		Group:    "user",
	}

	if err = config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError,
			Message: "注册用户失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    http.StatusOK, //200 ok
		Message: "注册用户成功",
		Data: gin.H{
			"username": newUser.Username,
			"user_id":  newUser.ID,
		},
	})
}

// @Summary		用户登录
// @Description	用户登录
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		models.LoginRequest	true	"注册请求"
// @Success		200		{object}	models.Response		"登录成功"
// @Failure		400		{object}	models.Response		"请求参数错误"
// @Failure		403		{object}	models.Response		"认证失败"
// @Failure		500		{object}	models.Response		"服务器错误"
// @Router			/api/auth/login [post]
func Login(c *gin.Context) {
	var req models.LoginRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    http.StatusBadRequest, //400
			Message: "请求参数错误",
		})
		return
	}

	var user models.User
	err = config.DB.Where("username = ?", req.Username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    http.StatusForbidden, //403
			Message: "用户不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "查询数据库失败：" + err.Error(),
		})
		return
	}

	err = utils.ComparePassword(user.Password, req.Password)
	if err != nil {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    http.StatusForbidden, //403
			Message: "密码错误",
		})
		return
	}

	//接下来获取所属权限组
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("group", user.Group)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "鉴权组件错误：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    http.StatusOK,
		Message: "登录成功",
		Data: gin.H{
			"user_id": user.ID,
			"group":   user.Group,
		},
	})
}

// @Summary		用户登出
// @Description	用户登出
// @Tags			auth
// @Security		ApikeyAuth
// @Produce		json
// @Success		200	{object}	models.Response	"登出成功"
// @Failure		500	{object}	models.Response	"服务器错误"
// @Router			/api/logout [post]
func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "登出失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    http.StatusOK,
		Message: "登出成功",
	})
}

//本来还想做个改密码的，发现有点难，蒜鸟蒜鸟

// @Summary		创建图书
// @Description	添加新图书(需要管理员权限)
// @Tags			books
// @Security		ApikeyAuth
// @Accept			multipart/form-data
// @Produce		json
// @Param title formData string ture "书名"
// @Param author formData string ture "作者"
// @Param summary formData string false "简介"
// @Param cover formData file false "封面图片"
// @Param initial_stock formData integer true "初始库存" minimum(0)
// @Success		200		{object}	models.Response{data=models.Book}		"创建成功"
// @Failure		400		{object}	models.Response		"请求参数错误"
// @Failure		409		{object}	models.Response		"图书已存在"
// @Failure		500		{object}	models.Response		"服务器错误"
// @Router			/api/books [post]
func CreateBook(c *gin.Context) {
	var req models.CreateBookRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    http.StatusBadRequest, //400
			Message: "请求参数错误",
		})
		return
	}

	var searchedBook models.Book
	err := config.DB.Where("title = ? AND author = ?", req.Title, req.Author).First(&searchedBook).Error
	if err == nil {
		c.JSON(http.StatusConflict, models.Response{
			Code:    http.StatusConflict, //409
			Message: "图书已存在(书名与作者相同)",
		})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "查询数据库失败：" + err.Error(),
		})
		return
	}

	coverPath := config.DefualtCoverPath
	if req.Cover != nil && req.Cover.Size > 0 {
		log.Printf("上传封面图片,Size:%d", req.Cover.Size)
		savePath, err := utils.SaveImages(c, req.Cover)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    http.StatusInternalServerError, //500
				Message: "图片保存失败：" + err.Error(),
			})
			return
		}
		coverPath = savePath
	} else {
		log.Printf("没有上传封面文件或者文件为空，使用默认路径：%s", coverPath)
	}
	summary := config.DefualtSummary
	if req.Summary != "" {
		summary = req.Summary
	}
	//简介就很简单了

	newBook := models.Book{
		Title:        req.Title,
		Author:       req.Author,
		Summary:      summary,
		CoverPath:    coverPath,
		InitialStock: req.InitialStock,
		Stock:        req.InitialStock,
		TotalStock:   req.InitialStock,
	}

	if err := config.DB.Create((&newBook)).Error; err != nil {
		//出错了顺便把封面也删了，别占地
		if req.Cover != nil && req.Cover.Size > 0 {
			if err := utils.RemoveFile(coverPath); err != nil {
				c.JSON(http.StatusInternalServerError, models.Response{
					Code:    http.StatusInternalServerError, //500
					Message: "删除封面失败（创建图书失败）：" + err.Error(),
				})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "创建图书失败（删除封面成功）：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    http.StatusOK,
		Message: "图书创建成功",
		Data:    newBook,
	})
}

// @Summary		删除图书
// @Description	通过book_id删除图书(需要管理员权限)
// @Tags			books
// @Security		ApikeyAuth
// @Produce		json
// @Param book_id path uint ture "图书ID"
// @Success		200		{object}	models.Response{data=models.Book}		"删除成功"
// @Failure		400		{object}	models.Response		"请求参数错误"
// @Failure		404		{object}	models.Response		"图书不存在"
// @Failure		409		{object}	models.Response		"图书借阅中"
// @Failure		500		{object}	models.Response		"服务器错误"
// @Router			/api/books/{id} [delete]
func DeletedBook(c *gin.Context) {
	id := c.Param("id")
	bookID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    http.StatusBadRequest, //400
			Message: "无效的图书ID",
		})
		return
	}

	req := models.FindBookRequest{ID: uint(bookID)}

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		//DB.Transaction是要么全都成功，要么全都失败的请求
		//其实像之前一样也可以，就是回传会比较分散
		//这也算是一种比较优美的方式？
		//tx是gorm给的变量名，是"transaction"(事务)的缩写
		//要说的话相当于操作表吧

		var searchedBook models.Book

		if err := tx.First(&searchedBook, req.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return config.ErrBookNotFound
			}
			return err
		}
		if searchedBook.Stock != searchedBook.TotalStock {
			return config.ErrBookBorrowed
		}

		if err := tx.Delete(&searchedBook).Error; err != nil {
			return config.ErrDeleteBook
		}
		if err := utils.RemoveFile(searchedBook.CoverPath); err != nil {
			return config.ErrDeleteCover
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, config.ErrBookNotFound) {
			c.JSON(http.StatusNotFound, models.Response{
				Code:    http.StatusNotFound, //404
				Message: "图书不存在",
			})
			return
		}
		if errors.Is(err, config.ErrDeleteBook) {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    http.StatusInternalServerError,
				Message: "图书删除失败：" + err.Error(),
			})
			return
		}
		if errors.Is(err, config.ErrDeleteCover) {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    http.StatusInternalServerError,
				Message: "封面删除失败：" + err.Error(),
			})
			return
		}
		if errors.Is(err, config.ErrBookBorrowed) {
			c.JSON(http.StatusConflict, models.Response{
				Code:    http.StatusConflict, //409
				Message: "图书出借中",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError,
			Message: "数据库意外错误：" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, models.Response{
		Code:    http.StatusOK,
		Message: "删除图书成功",
	})
}

// @Summary 获取图书列表
// @Description 按条件分页查询图书(顺序书名作者和简介)。 注意：
// @Description 1.默认每条最多30字。（主要是针对于简介）
// @Description 2.不能单独使用通配符（%和_，简单理解为mysql的正则表达式就ok），否则清空搜索。
// @Description 3.如果有%和_的查询会转义。
// @Description 4.实际效果如果三个参数都传，是返回符合这三个效果
// @Description 5.按更新时间倒序（最近修改的书在前）
// @Description 6.如果查询结果为空，返回状态码也是200，但是data为空
// @Tags books
// @Security ApiKeyAuth
// @Produce json
// @Param title query string false "按书名模糊查询"
// @Param author query string false "按作者模糊查询"
// @Param summary query string false "按简介模糊查询"
// @Success 200 {object} models.Response{data=[]models.Book} "查询成功"
// @Failure 500 {object} models.Response "数据库查询失败"
// @Router /api/books [get]
func GetBooks(c *gin.Context) {
	//搜索的话用mysql自带的模糊搜索就ok了
	var books []models.Book
	var err error

	// //分页逻辑
	// //默认在第一页
	// page := 1
	// if p := c.Query("page"); p != "" {
	// 	page, err = strconv.Atoi(p)
	// 	if err != nil {
	// 		c.JSON(http.StatusInternalServerError, models.Response{
	// 			Code:    http.StatusInternalServerError, //500
	// 			Message: "页码转换失败（或页码不符规范）",
	// 		})
	// 		return
	// 	}
	// 	if page < 1 {
	// 		page = 1
	// 	}
	// 	//若页码小于1，强制转化为1
	// }
	// query := config.DB.Model(&models.Book{})
	// query.Limit(config.DefualtGetBooksQueryLimit).Offset((page - 1) * config.DefualtGetBooksQueryLimit)

	title := utils.SqlSafeLikeKeyword(c.Query(("title")))
	author := utils.SqlSafeLikeKeyword(c.Query(("author")))
	summary := utils.SqlSafeLikeKeyword(c.Query(("summary")))
	//直接用%xxx%了，小项目，懒得优化了也，但是不让用户用通配符捏

	// // 默认50,分页，防止轰炸
	// // 本来想写在config的，但是go不能循环import
	//分页弃掉，懒得做了（
	query := config.DB.Model(&models.Book{})

	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if author != "" {
		query = query.Where("author LIKE ?", "%"+author+"%")
	}
	if summary != "" {
		query = query.Where("summary LIKE ?", "%"+summary+"%")
	}

	//没用order的话是无序的，所以干脆按照修改时间排序了
	if err = query.Order("updated_at DESC").Find(&books).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError,
			Message: "数据库查询失败：" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, models.Response{
		Code:    http.StatusOK, //200
		Message: "查询成功",
		Data:    books,
	})
}

// @Summary 更新图书
// @Description 通过book_id修改图书信息（管理员权限）
// @Tags books
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param book_id path uint true "图书ID"
// @Param title formData string false "新书名"
// @Param author formData string false "新作者"
// @Param summary formData string false "新简介"
// @Param cover formData file false "新封面"
// @Param stock formData integer false "当前库存" minimum(0)
// @Param total_stock formData integer false "总库存" minimum(0)
// @Success 200 {object} models.Response{data=models.Book} "更新成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 404 {object} models.Response "图书不存在"
// @Failure 500 {object} models.Response "服务器错误"
// @Router /api/books/{id} [put]
func UpdateBook(c *gin.Context) {
	id := c.Param("book_id")
	bookID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    http.StatusBadRequest, //400
			Message: "无效的图书ID",
		})
		return
	}

	var req models.UpdateBookRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    http.StatusBadRequest, //400
			Message: "参数设定错误",
		})
		return
	}
	req.ID = uint(bookID)

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	//这里是开了一个业务表
	//emmm我也不知道有没有什么通用的叫法，我这么叫
	//他这个表的作用相当于是说记录步骤，如果说出错了我可以回滚
	//然后最后没问题的话就提交
	//保证操作完整性说是
	//一次性更新完所有东西
	//它跟新建书不一样
	//recover的话也能捕捉程序崩溃
	//嗯就大概是这样
	//也算是多种写法吧

	var upBook models.Book
	result := tx.Set("gorm:query_option", "FOR UPDATE").First(&upBook, req.ID)
	//FOR UPDATE是SQL锁机制
	//即使还没做啥操作，也要在错误后解除锁定再return
	err = result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, models.Response{
				Code:    http.StatusNotFound, //404
				Message: "图书不存在",
			})
		} else {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    http.StatusInternalServerError, //500
				Message: "数据库查询失败：" + err.Error(),
			})
		}
		return
	}

	updates := make(map[string]interface{})
	//gorm的update要求用map作为参数
	//其实也合理，map也确实方便
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Author != "" {
		updates["author"] = req.Author
	}
	if req.Summary != "" {
		updates["summary"] = req.Summary
	}

	//然后是stock和totalstock的逻辑
	//其实本身就一点，你总量不能小于存量
	if req.Stock >= 0 || req.TotalStock > 0 {
		newStock := upBook.Stock
		newTotalStock := upBook.TotalStock
		if req.Stock >= 0 {
			newStock = req.Stock
		}
		if req.TotalStock >= 0 {
			newTotalStock = req.TotalStock
		}

		if newStock > newTotalStock {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    http.StatusBadRequest, //400
				Message: "当前库存不得大于总库存",
			})
			return
		}

		updates["stock"] = newStock
		updates["total_stock"] = newTotalStock
	}

	//接下来处理cover逻辑
	//从Create那里搬过来一点
	coverPath := config.DefualtCoverPath
	if req.Cover != nil && req.Cover.Size > 0 {
		log.Printf("上传封面图片,Size:%d", req.Cover.Size)
		savePath, err := utils.SaveImages(c, req.Cover)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    http.StatusInternalServerError, //500
				Message: "图片保存失败：" + err.Error(),
			})
			return
		}
		coverPath = savePath
	} else {
		log.Printf("没有上传封面文件或者文件为空，使用默认路径：%s", coverPath)
	}
	//这就ok了，跟Create一摸一样
	//我说代码复用
	updates["cover_path"] = coverPath

	//这里一开始是先cover再stock
	//但我后来换了一下，因为如果说stock错误，我就不应该去下这个封面了
	//省时并且省力

	if len(updates) > 0 {
		if err = tx.Model(&upBook).Updates(updates).Error; err != nil {
			tx.Rollback()
			if req.Cover != nil && req.Cover.Size > 0 {
				if err := utils.RemoveFile(coverPath); err != nil {
					c.JSON(http.StatusInternalServerError, models.Response{
						Code:    http.StatusInternalServerError, //500
						Message: "删除封面失败（更新失败）：" + err.Error(),
					})
					return
				}
				c.JSON(http.StatusInternalServerError, models.Response{
					Code:    http.StatusInternalServerError, //500
					Message: "更新失败（删除封面成功）：" + err.Error(),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    http.StatusInternalServerError, //500
				Message: "更新失败：" + err.Error(),
			})
			return
		}
	}
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		if req.Cover != nil && req.Cover.Size > 0 {
			if err := utils.RemoveFile(coverPath); err != nil {
				c.JSON(http.StatusInternalServerError, models.Response{
					Code:    http.StatusInternalServerError, //500
					Message: "删除封面失败（提交失败）：" + err.Error(),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    http.StatusInternalServerError, //500
				Message: "提交失败（删除封面成功）：" + err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "提交失败：" + err.Error(),
		})
		return
	}
	var book models.Book
	config.DB.First(&book, req.ID)

	c.JSON(http.StatusOK, models.Response{
		Code:    http.StatusOK, //200
		Message: "书籍更新成功",
		Data:    book,
	})
}
