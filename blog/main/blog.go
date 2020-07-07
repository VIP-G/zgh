package main

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql" // 导入数据库驱动
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"time"
)

type User struct {
	Id         int `orm:"auto"` //如果表的主键不是id，那么需要加上pk注释，显式的说这个字段是主键
	Username   string
	Password   string
	Createtime string
	Article    []*Article `orm:"reverse(many)"` // 设置一对多的反向关系    一方
	Comment    []*Comment `orm:"reverse(many)"` // 设置一对多的反向关系

}
type Article struct {
	Id      int `orm:"auto"`
	Title   string
	Content string
	Pubtime string
	User    *User      `orm:"rel(fk)"`       // 设置一对多关系
	Comment []*Comment `orm:"reverse(many)"` // 设置一对多的反向关系
}
type Comment struct {
	Id      int `orm:"auto"`
	Content string
	User    *User    `orm:"rel(fk)"`
	Article *Article `orm:"rel(fk)"` // 设置一对多关系   多方
}

func init() {

	// 设置默认数据库
	orm.RegisterDataBase("default", "mysql", "root:123456@/blog_data?charset=utf8", 30)

	// 注册定义的 model
	orm.RegisterModel(new(User), new(Article), new(Comment))
	// RegisterModel 也可以同时注册多个 model
	// orm.RegisterModel(new(User), new(Profile), new(Post))

	// 创建 table
	orm.RunSyncdb("default", false, true)
	orm.SetMaxIdleConns("default", 30)

}
func change_str(x []string) string {
	str := ""
	for _, v := range x {
		str += v
	}
	return str
}
func login(w http.ResponseWriter, r *http.Request) {
	o := orm.NewOrm()
	fmt.Println("method:", r.Method)
	if r.Method == "GET" {

		t, _ := template.ParseFiles("login.gtpl")
		log.Println(t.Execute(w, nil))

	} else {

		err := r.ParseForm() // 解析 url 传递的参数，对于 POST 则解析响应包的主体（request body）
		if err != nil {
			log.Fatal("ParseForm: ", err)
		}

		us_name_str := change_str(r.Form["username"])

		us_pwd_str := change_str(r.Form["password"])

		// 查询是否存在用户，存在校验密码
		var list []orm.Params
		num, err := o.Raw("SELECT * FROM user WHERE username = ?", us_name_str).Values(&list)
		if err != nil {
			fmt.Fprintf(w, "未知错误，请重试")
		} else {
			if num == 0 {
				fmt.Fprintf(w, "用户不存在，请先注册")
			} else {
				// fmt.Println("用户密码",list[0]["password"])
				user_pwd := list[0]["password"]
				if us_pwd_str == user_pwd {
					fmt.Fprintf(w, "登录成功")
				} else {
					fmt.Fprintf(w, "密码错误")
				}

			}

		}

	}
}
func register(w http.ResponseWriter, r *http.Request) {
	o := orm.NewOrm()
	if r.Method == "GET" {
		t, _ := template.ParseFiles("register.gtpl")
		log.Println(t.Execute(w, nil))
	} else {
		err := r.ParseForm() // 解析 url 传递的参数，对于 POST 则解析响应包的主体（request body）
		if err != nil {
			log.Fatal("ParseForm: ", err)
		}

		pwd := r.Form["password"]
		pwd2 := r.Form["password2"]
		us_name_str := change_str(r.Form["username"])

		us_pwd_str := ""
		// 查询用户是否存在，不存在创建用户
		var list []orm.Params
		num, err := o.Raw("SELECT * FROM user WHERE username = ?", us_name_str).Values(&list)
		if err != nil {
			fmt.Fprintf(w, "未知错误，请重试")
		} else {
			if num == 0 {

				if len(pwd) == len(pwd2) {
					for i, v := range pwd {
						if v == pwd2[i] {
							us_pwd_str += v

							// 插入表数据 
							var user User
							user.Username = us_name_str
							user.Password = us_pwd_str
							user.Createtime = time.Now().Format("2006-01-01 15:04:05")
							_, err := o.Insert(&user)
							if err == nil {
								fmt.Fprintf(w, "注册成功，去登陆")
							} else {
								fmt.Fprintf(w, "未知错误，请重试")
							}
						} else {

							fmt.Fprintf(w, "密码不一致，重新输入")
						}
					}
				} else {

					fmt.Fprintf(w, "密码不一致，重新输入")
				}
			} else {
				fmt.Fprintf(w, "用户名已存在")
			}

		}

	}
}
func writearticle(w http.ResponseWriter, r *http.Request) {
	o := orm.NewOrm()
	if r.Method == "GET" {
		t, _ := template.ParseFiles("writearticle.gtpl")
		log.Println(t.Execute(w, nil))
	} else {
		err := r.ParseForm() // 解析 url 传递的参数，对于 POST 则解析响应包的主体（request body）
		if err != nil {
			log.Fatal("ParseForm: ", err)
		}
		wri_title := change_str(r.Form["title"])
		wri_content := change_str(r.Form["content"])
		wri_user := change_str(r.Form["username"])

		//fmt.Println(r.Form["title"], wri_title)
		//fmt.Println(r.Form["content"], wri_content)
		//fmt.Println(r.Form["username"], wri_user)

		var list []orm.Params
		num, err := o.Raw("SELECT * FROM user WHERE username = ?", wri_user).Values(&list)
		if err != nil {
			fmt.Fprintf(w, "未知错误，请重试")
		} else {
			if num == 0 {
				fmt.Fprintf(w, "当前用户不存在，请输入正确用户名")
			} else {
				arti_title := Article{Title: wri_title}
				er := o.Read(&arti_title, "Title")
				if er == orm.ErrNoRows {
					//fmt.Println("用户信息", list[0]["id"])
					x := list[0]["id"]
					n1, _ := strconv.Atoi(x.(string))
					//插入文章信息
					var article Article
					article.Content = wri_content
					article.Title = wri_title
					article.Pubtime = time.Now().Format("2006-01-01 15:04:05")
					u := User{Id: n1}
					article.User = &u
					_, err := o.Insert(&article)
					if err == nil {
						fmt.Fprintf(w, "发表成功")
					} else {
						fmt.Fprintf(w, "未知错误，请重试")
					}
				} else {
					fmt.Fprintf(w, "标题重复")
				}

			}

		}

	}
}
func comment(w http.ResponseWriter, r *http.Request) {
	o := orm.NewOrm()
	if r.Method == "GET" {
		t, _ := template.ParseFiles("comment.gtpl")
		log.Println(t.Execute(w, nil))
	} else {
		err := r.ParseForm() // 解析 url 传递的参数，对于 POST 则解析响应包的主体（request body）
		if err != nil {
			log.Fatal("ParseForm: ", err)
		}
		comment_article := change_str(r.Form["article"])
		comment_content := change_str(r.Form["content"])
		comment_user := change_str(r.Form["username"])
		//查询用户和文章是否存在
		user := User{Username: comment_user}
		err2 := o.Read(&user, "Username")
		if err2 == orm.ErrNoRows {
			fmt.Fprintf(w, "当前用户不存在，请输入正确用户名")
		} else {
			arti := Article{Title: comment_article}
			err3 := o.Read(&arti, "Title")
			if err3 == orm.ErrNoRows {
				fmt.Fprintf(w, "当前文章不存在")
			} else {
				var comment Comment
				comment.Content = comment_content
				comment.User = &user
				comment.Article = &arti
				_, err := o.Insert(&comment)
				if err == nil {
					fmt.Fprintf(w, "发表成功")
				} else {
					fmt.Fprintf(w, "未知错误，请重试")
				}
			}
		}
	}

}
func getarticle(w http.ResponseWriter, r *http.Request) {
	o := orm.NewOrm()
	if r.Method == "GET" {
		//原生sql以其他字段为查询条件
		var list []orm.Params
		num, err := o.Raw("SELECT * FROM article").Values(&list)
		if err != nil {
			fmt.Println(err)
		} else {
			if num == 0 {
				fmt.Fprintf(w, "没有文章")
			} else {

				fmt.Fprintf(w, "", list)
				fmt.Println(list)
			}
		}
	}

}
func getcomment(w http.ResponseWriter, r *http.Request) {
	o := orm.NewOrm()
	if r.Method == "GET" {
		t, _ := template.ParseFiles("getcomment.gtpl")
		log.Println(t.Execute(w, nil))
	} else {
		err := r.ParseForm()
		if err != nil {
			log.Fatal("ParseForm: ", err)
		}
		article_title := change_str(r.Form["title"])
		arti := Article{Title: article_title}
		err2 := o.Read(&arti, "Title")
		if err2 == orm.ErrNoRows {
			fmt.Fprintf(w, "当前文章不存在")
		} else {
			fmt.Println(arti.Id, "请论文章id")
			var list []orm.Params
			num, err := o.Raw("SELECT * FROM comment WHERE article_id = ?", arti.Id).Values(&list)
			if err != nil {

				fmt.Fprintf(w, "读取评论失败")
			} else {
				if num == 0 {
					fmt.Fprintf(w, "该文章暂无评论，去评论")

				} else {
					fmt.Fprintf(w, "", list)
				}
			}
		}

	}

}
func main() {
	orm.Debug = true
	http.HandleFunc("/login", login) // 设置访问的路由
	http.HandleFunc("/register", register)
	http.HandleFunc("/writearticle", writearticle)
	http.HandleFunc("/getarticle", getarticle)
	http.HandleFunc("/comment", comment)
	http.HandleFunc("/getcomment", getcomment)

	err := http.ListenAndServe(":9090", nil) // 设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	var w io.Writer
	orm.DebugLog = orm.NewLog(w)

}
