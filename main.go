package main

import (
	"fmt"
	"math"
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
	ID   int    `gorm:"primary_key"`
	Name string `gorm:"NOT NULL"`
}

// TableName Categoryのデフォルトテーブル名変更
// func (o *Category) TableName() string {
// 	return "category"
// }

// Memo 構造体
type Memo struct {
	ID         int `gorm:"primary_key"`
	Category   Category
	CategoryID int    `gorm:"NOT NULL"`
	Title      string `gorm:"NOT NULL"`
	Detail     string `gorm:"DEFAULT:''"`
	Attention  int    `gorm:"DEFAULT:0"`
}

// TableName Memoのデフォルトテーブル名変更
// func (o *Memo) TableName() string {
// 	return "memo"
// }

// 複数テンプレート用 これを使用　→　https://github.com/gin-contrib/multitemplate
func createMyRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("index", "templates/base.html", "templates/index.html")
	r.AddFromFiles("detail", "templates/base.html", "templates/detail.html")
	r.AddFromFiles("detail_form", "templates/base.html", "templates/detail_form.html")
	r.AddFromFiles("insert", "templates/base.html", "templates/insert.html")
	r.AddFromFiles("category", "templates/base.html", "templates/category.html")
	return r
}

// Paginate 構造体
type Paginate struct {
	Pages       []uint //数字で回数を指定して繰り返し処理ができないので配列にする
	CurrentPage uint
	Limit       uint
	Offset      uint
}

// 簡易ページネーション
func pagination(current uint, total uint) *Paginate {
	//総ページ数 = Ceil(総件数÷1ページ表示数)

	//1ページ表示数
	var number uint = 50

	//総ページ数（数字で回数を指定して繰り返し処理ができないので配列にする）
	pages := int(math.Ceil(float64(total) / float64(number)))
	var pagesArr []uint
	for i := 1; i <= pages; i++ {
		pagesArr = append(pagesArr, uint(i))
	}

	//構造体生成
	a := Paginate{
		Pages:       pagesArr,
		CurrentPage: current,
		Limit:       uint(number),
		Offset:      current*number - number,
	}
	return &a
}

// index
func index(c *gin.Context) {
	db := DB()
	defer db.Close()

	//現ページ取得
	page := 1
	if c.Query("page") != "" {
		var err error
		page, err = strconv.Atoi(c.Query("page"))
		if err != nil {
			c.String(400, "400 Bad Request")
			return
		}
	}

	//Category 複数レコード抽出
	cates := []Category{}
	db.Order("name").Find(&cates)

	//Memo 複数レコード抽出
	memos := []Memo{}

	//件数取得
	count := 0
	db.Find(&memos).Count(&count)

	//Paginate取得
	var pageObj = pagination(uint(page), uint(count))

	//データ取得
	db.Preload("Category").Limit(pageObj.Limit).Offset(pageObj.Offset).Order("id desc").Find(&memos)

	//テンプレートに変数を渡す
	c.HTML(http.StatusOK, "index", gin.H{
		"title":      "トップページ",
		"cateno":     0,
		"cates":      cates,
		"memos":      memos,
		"top":        true,
		"pagination": pageObj,
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

	//HTMLのエスケープをしておく（全部マークダウンにするつもりなのでコメントアウト）
	//memo.Detail = html.EscapeString(memo.Detail)

	//10倍にして返す
	memo.Attention = memo.Attention * 10

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

	//0-10までの配列生成
	attentionArr := make([]int, 11)

	//テンプレートに変数を渡す
	c.HTML(http.StatusOK, "detail_form", gin.H{
		"title":     memo.Title,
		"cateno":    0,
		"cates":     cates,
		"memo":      memo,
		"attention": attentionArr,
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
	attention, _ := strconv.Atoi(c.PostForm("attention"))

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
		memo.CategoryID = categoryID
		memo.Title = title
		memo.Detail = detail
		memo.Attention = attention

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

	//0-10までの配列生成
	attentionArr := make([]int, 11)

	//テンプレートに変数を渡す
	c.HTML(http.StatusOK, "insert", gin.H{
		"title":     "新規追加",
		"cateno":    0,
		"cates":     cates,
		"attention": attentionArr,
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
	attention, _ := strconv.Atoi(c.PostForm("attention"))

	//Insert
	obj := Memo{
		CategoryID: categoryID,
		Title:      title,
		Detail:     detail,
		Attention:  attention,
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
		"title":  "カテゴリ編集",
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
