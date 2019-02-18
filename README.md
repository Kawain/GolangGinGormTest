# GolangGinGormTest

既存のSQLiteのデータベースファイルがある前提で作りました。  
以下の3つを利用しています。

`go get -u github.com/gin-gonic/gin`

`go get github.com/gin-contrib/multitemplate`

`go get -u github.com/jinzhu/gorm`

## Golang

サクッと動くので気持ちが良い…。  
Gormを使ったが生のSQLでもいいと思う。というかGormの使い方がまだよくわからない。  

## marked.min.js

marked.min.js を利用してマークダウン記法対応しようと思った。  
しかし、訳あってpreで表示しているために、うまくいかない。  
最初からマークダウンで書く場合はpreタグを変えたほうがいい。  
コードは html.EscapeString してからバッククォート3つではおかしくなる。  

## ハマったこと

テンプレートのループ内でループ外の変数を使うとき$(ドルマーク)を先頭につけて利用するのを知らなかった

```
<select class="form-control" id="exampleFormControlSelect1">
{{range .cates}}
  {{ if eq .ID $.memo.Category.ID }}
  <option value="{{.ID}}" selected>{{.Name}}</option>
  {{else}}
  <option value="{{.ID}}">{{.Name}}</option>
  {{end}}
{{end}}
</select>
```

テンプレートでコレクションのループはできるが、1～10までとかの数値でのループができないので困った。  
結局配列を作ってテンプレートに渡した。  

コンパイルしたら28.2 MBになった。
