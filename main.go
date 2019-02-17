package main

import (
	"fmt"
	"html"
	"net/http"
	"strconv"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// DB 接続
func DB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "notes.db")
	if err != nil {
		panic("データベース接続失敗")
	}
	db.AutoMigrate(&Category{}, &Memo{})
	return db
}

// Category 構造体
type Category struct {
	ID   uint `gorm:"primary_key"`
	Name string
}

// TableName Categoryのデフォルトテーブル名変更
func (o *Category) TableName() string {
	return "category"
}

// Memo 構造体
type Memo struct {
	ID         uint `gorm:"primary_key"`
	Category   Category
	CategoryID uint
	Title      string
	Detail     string
}

// TableName Memoのデフォルトテーブル名変更
func (o *Memo) TableName() string {
	return "memo"
}

// 複数テンプレート用 これを使用　→　https://github.com/gin-contrib/multitemplate
func createMyRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("index", "templates/base.html", "templates/index.html")
	r.AddFromFiles("detail", "templates/base.html", "templates/detail.html")
	r.AddFromFiles("detail_form", "templates/base.html", "templates/detail_form.html")
	r.AddFromFiles("insert", "templates/base.html", "templates/insert.html")
	r.AddFromFiles("category", "templates/base.html", "templates/category.html")
	r.AddFromFiles("about", "templates/base.html", "templates/about.html")
	r.AddFromFiles("contact", "templates/base.html", "templates/contact.html")
	return r
}

// index
func index(c *gin.Context) {
	db := DB()
	defer db.Close()

	//Category 複数レコード抽出
	cates := []Category{}
	db.Order("name").Find(&cates)

	//Memo 複数レコード抽出
	memos := []Memo{}
	db.Preload("Category").Order("id desc").Find(&memos)

	//テンプレートに変数を渡す
	c.HTML(http.StatusOK, "index", gin.H{
		"title":  "トップページ",
		"cateno": 0,
		"cates":  cates,
		"memos":  memos,
		"top":    true,
	})
}

// search
func search(c *gin.Context) {
	db := DB()
	defer db.Close()

	//数値に変換
	ID, _ := strconv.Atoi(c.Param("id"))
	q := c.Query("q")

	//Category 複数レコード抽出
	cates := []Category{}
	db.Order("name").Find(&cates)

	//Memo レコード 複数レコード抽出
	memos := []Memo{}
	//db.Preload("Category").Where("category_id = ?", ID).Order("id desc").Find(&memos)
	if ID == 0 {
		db.Preload("Category").
			Where("title LIKE ?", "%"+q+"%").
			Or("detail LIKE ?", "%"+q+"%").
			Order("id desc").
			Find(&memos)
	} else {
		db.Preload("Category").
			Where("category_id = ? AND title LIKE ?", ID, "%"+q+"%").
			Or("category_id = ? AND detail LIKE ?", ID, "%"+q+"%").
			Order("id desc").
			Find(&memos)
	}

	//なければ404
	if len(memos) == 0 {
		c.String(404, "404 page not found")
		return
	}

	//タイトルの作成
	var title string

	if ID > 0 && len(q) > 0 {
		title = memos[0].Category.Name + " " + q
	} else if ID > 0 {
		title = memos[0].Category.Name
	} else if len(q) > 0 {
		title = q
	}

	//テンプレートに変数を渡す
	c.HTML(http.StatusOK, "index", gin.H{
		"title":  title,
		"cateno": ID,
		"cates":  cates,
		"memos":  memos,
		"top":    false,
	})
}

// detail
func detail(c *gin.Context) {
	db := DB()
	defer db.Close()

	//数値に変換
	ID, _ := strconv.Atoi(c.Param("id"))

	//Category 複数レコード抽出
	cates := []Category{}
	db.Order("name").Find(&cates)

	//Memo 単一レコード
	memo := Memo{}
	db.Preload("Category").Where("id = ?", ID).First(&memo)

	//なければ404
	if memo.ID == 0 {
		c.String(404, "404 page not found")
		return
	}

	//HTMLのエスケープをしておく
	memo.Detail = html.EscapeString(memo.Detail)

	//テンプレートに変数を渡す
	c.HTML(http.StatusOK, "detail", gin.H{
		"title":  memo.Title,
		"cateno": 0,
		"cates":  cates,
		"memo":   memo,
	})
}

// detailForm
func detailForm(c *gin.Context) {
	db := DB()
	defer db.Close()

	//数値に変換
	ID, _ := strconv.Atoi(c.Param("id"))

	//Category 複数レコード抽出
	cates := []Category{}
	db.Order("name").Find(&cates)

	//Memo 単一レコード
	memo := Memo{}
	db.Preload("Category").Where("id = ?", ID).First(&memo)

	//なければ404
	if memo.ID == 0 {
		c.String(404, "404 page not found")
		return
	}

	//テンプレートに変数を渡す
	c.HTML(http.StatusOK, "detail_form", gin.H{
		"title":  memo.Title,
		"cateno": 0,
		"cates":  cates,
		"memo":   memo,
	})
}

// detailExecut 更新か削除
func detailExecut(c *gin.Context) {
	db := DB()
	defer db.Close()

	//数値に変換
	ID, _ := strconv.Atoi(c.PostForm("id"))
	categoryID, _ := strconv.Atoi(c.PostForm("category_id"))

	title := c.PostForm("title")
	detail := c.PostForm("detail")
	delete := c.PostForm("delete")
	update := c.PostForm("update")

	if delete == "delete" {
		//トランザクション
		tx := db.Begin()

		//update
		if err := tx.Where("id = ?", uint(ID)).Delete(Memo{}).Error; err != nil {
			tx.Rollback()
			c.String(500, "削除に失敗しました")
			return
		}

		tx.Commit()

		//リダイレクト
		c.Redirect(http.StatusMovedPermanently, "/")

	} else if update == "update" {

		//Memo 単一レコード
		memo := Memo{}
		db.Where("id = ?", ID).First(&memo)
		memo.CategoryID = uint(categoryID)
		memo.Title = title
		memo.Detail = detail

		//トランザクション
		tx := db.Begin()

		//update
		if err := tx.Save(&memo).Error; err != nil {
			tx.Rollback()
			c.String(500, "修正に失敗しました")
			return
		}

		tx.Commit()

		//リダイレクト
		c.Redirect(http.StatusMovedPermanently, "/detail/"+fmt.Sprint(ID))

	} else {
		c.String(400, "400 Bad Request")
	}
}

// insert フォーム
func insert(c *gin.Context) {
	db := DB()
	defer db.Close()

	//Category 複数レコード抽出
	cates := []Category{}
	db.Order("name").Find(&cates)

	//テンプレートに変数を渡す
	c.HTML(http.StatusOK, "insert", gin.H{
		"title":  "新規追加",
		"cateno": 0,
		"cates":  cates,
	})
}

// insertExecut 実行
func insertExecut(c *gin.Context) {
	db := DB()
	defer db.Close()

	//数値に変換
	categoryID, _ := strconv.Atoi(c.PostForm("category_id"))
	title := c.PostForm("title")
	detail := c.PostForm("detail")

	//Insert
	obj := Memo{
		CategoryID: uint(categoryID),
		Title:      title,
		Detail:     detail,
	}
	db.Create(&obj)

	//リダイレクト
	c.Redirect(http.StatusMovedPermanently, "/")
}

// category 編集フォーム
func category(c *gin.Context) {
	db := DB()
	defer db.Close()

	//Category 複数レコード抽出
	cates := []Category{}
	db.Order("name").Find(&cates)

	//ソートが違うカテゴリも抽出
	cates2 := []Category{}
	db.Order("id").Find(&cates2)

	//テンプレートに変数を渡す
	c.HTML(http.StatusOK, "category", gin.H{
		"title":  "カテゴリ編",
		"cateno": 0,
		"cates":  cates,
		"cates2": cates2,
	})
}

// category 追加
func categoryInsert(c *gin.Context) {
	db := DB()
	defer db.Close()

	//POSTしたform input nameを取得
	name := c.PostForm("name")

	//トランザクション
	tx := db.Begin()

	//Insert
	if err := tx.Create(&Category{Name: name}).Error; err != nil {
		tx.Rollback()
		c.String(500, "追加に失敗しました")
		return
	}

	tx.Commit()

	//リダイレクト
	c.Redirect(http.StatusMovedPermanently, "/category")
}

// category 修正
func categoryUpdate(c *gin.Context) {
	db := DB()
	defer db.Close()

	//数値に変換
	ID, _ := strconv.Atoi(c.PostForm("id"))
	//POSTしたform input nameを取得
	name := c.PostForm("name")

	//トランザクション
	tx := db.Begin()

	//update
	if err := tx.Model(Category{}).Where("id = ?", ID).Update("name", name).Error; err != nil {
		tx.Rollback()
		c.String(500, "修正に失敗しました")
		return
	}

	tx.Commit()

	//リダイレクト
	c.Redirect(http.StatusMovedPermanently, "/category")
}

// category 削除
func categoryDelete(c *gin.Context) {
	db := DB()
	defer db.Close()

	//数値に変換
	ID, _ := strconv.Atoi(c.PostForm("id"))

	//トランザクション
	tx := db.Begin()

	//update
	if err := tx.Where("id = ?", uint(ID)).Delete(Category{}).Error; err != nil {
		tx.Rollback()
		c.String(500, "削除に失敗しました")
		return
	}

	tx.Commit()

	//リダイレクト
	c.Redirect(http.StatusMovedPermanently, "/category")
}

func main() {
	router := gin.Default()
	router.HTMLRender = createMyRender()
	//静的ファイル
	router.Static("/assets", "./assets")
	//トップページ
	router.GET("/", index)
	//検索+カテゴリ
	router.GET("/category/:id", search)
	//詳細ページ
	router.GET("/detail/:id", detail)
	//詳細ページ編集フォーム
	router.GET("/detail/:id/form", detailForm)
	router.POST("/detail_execut", detailExecut)
	//新規追加
	router.GET("/insert", insert)
	router.POST("/insert_execut", insertExecut)
	//カテゴリ編集
	router.GET("/category", category)
	router.POST("/category_insert", categoryInsert)
	router.POST("/category_update", categoryUpdate)
	router.POST("/category_delete", categoryDelete)

	router.Run()
}
