package main

import (
	// "github.com/3xxx/engineercms/commands"
	"github.com/3xxx/engineercms/commands/daemon"
	"github.com/3xxx/engineercms/conf"
	"github.com/3xxx/engineercms/controllers"
	_ "github.com/3xxx/engineercms/controllers/version"
	"github.com/3xxx/engineercms/models"
	_ "github.com/3xxx/engineercms/routers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/astaxie/beego/session/memcache"
	_ "github.com/astaxie/beego/session/mysql"
	_ "github.com/astaxie/beego/session/redis"
	"github.com/astaxie/beego/toolbox"
	"github.com/kardianos/service"
	_ "github.com/mattn/go-sqlite3"
	// "github.com/go-xorm/xorm"
	// "github.com/astaxie/beego/plugins/cors"

	// _ "github.com/mattn/go-sqlite3"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
	// _ "github.com/3xxx/engineercms/routers"
)

func main() {
	// beego.AddFuncMap("dict", dict)
	beego.AddFuncMap("loadtimes", loadtimes)
	beego.AddFuncMap("subsuffix", subsuffix)
	//默认关闭orm调试模式
	ormDebug := beego.AppConfig.String("ormDebug")
	if ormDebug == "true" {
		orm.Debug = true
	} else {
		orm.Debug = false
	}
	//创建附件目录ModePerm FileMode = 0777 // 覆盖所有Unix权限位（用于通过&获取类型位）
	os.Mkdir("attachment", os.ModePerm)
	//创建轮播图片存放目录
	os.Mkdir("attachment//carousel", os.ModePerm)
	//自动建表
	orm.RunSyncdb("default", false, true)
	models.InsertUser()
	// insertGroup()
	// models.InsertRole()
	// time1 := "0/" + time + " * * * * *"

	time1 := "* 30 8 * * 1-5"
	//"* 30 8 * * 1-5"
	tk1 := toolbox.NewTask("tk1", time1, func() error { controllers.SendMessage(); return nil }) //func() error { fmt.Println("tk1"); return nil }
	toolbox.AddTask("tk1", tk1)
	toolbox.StartTask()
	defer toolbox.StopTask()

	// ********mindoc*********
	// if len(os.Args) >= 3 && os.Args[1] == "service" {
	// 	if os.Args[2] == "install" {
	//    daemon.Install()
	// 	} else if os.Args[2] == "remove" {
	// 		daemon.Uninstall()
	// 	} else if os.Args[2] == "restart" {
	// 		daemon.Restart()
	// 	}
	// }

	initialization()
	// commands.RegisterCache()
	// commands.RegisterLogger(conf.LogFile)

	// commands.RegisterCommand()

	d := daemon.NewDaemon()

	s, err := service.New(d, d.Config())

	if err != nil {
		fmt.Println("Create service error => ", err)
		os.Exit(1)
	}

	if err := s.Run(); err != nil {
		log.Fatal("启动程序失败 ->", err)
	}
	// ********mindoc*********

	beego.Run()
}

//显示页面加载时间
func loadtimes(t time.Time) int {
	return int(time.Now().Sub(t).Nanoseconds() / 1e6)
}

//去除扩展名
func subsuffix(in string) string {
	fileSuffix := path.Ext(in)
	return strings.TrimSuffix(in, fileSuffix)
}

// GoLang 如何在网页显示当前环境的版本号
// func main() {
// 	server()
// }
func server() {
	http.HandleFunc("/version", version)
	http.ListenAndServe(":8080", nil)
}

func version(w http.ResponseWriter, r *http.Request) {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w, fmt.Sprintf("%s", out))
}

//初始化数据
func initialization() {

	err := models.NewOption().Init()

	if err != nil {
		panic(err.Error())
		os.Exit(1)
	}

	member, err := models.NewMember().FindByFieldFirst("account", "admin")
	if err == orm.ErrNoRows {

		member.Account = "admin"
		member.Avatar = conf.URLForWithCdnImage("/static/mindoc/images/headimgurl.jpg")
		member.Password = "123456"
		member.AuthMethod = "local"
		member.Role = 0
		member.Email = "admin@iminho.me"

		if err := member.Add(); err != nil {
			panic("Member.Add => " + err.Error())
			os.Exit(0)
		}

		book := models.NewBook()

		book.MemberId = member.MemberId
		book.BookName = "MinDoc演示项目"
		book.Status = 0
		book.ItemId = 1
		book.Description = "这是一个MinDoc演示项目，该项目是由系统初始化时自动创建。"
		book.CommentCount = 0
		book.PrivatelyOwned = 0
		book.CommentStatus = "closed"
		book.Identify = "mindoc"
		book.DocCount = 0
		book.CommentCount = 0
		book.Version = time.Now().Unix()
		book.Cover = conf.GetDefaultCover()
		book.Editor = "markdown"
		book.Theme = "default"

		if err := book.Insert(); err != nil {
			panic("初始化项目失败 -> " + err.Error())
			os.Exit(1)
		}
	}

	if !models.NewItemsets().Exist(1) {
		item := models.NewItemsets()
		item.ItemName = "默认项目空间"
		item.MemberId = 1
		if err := item.Save(); err != nil {
			panic("初始化项目空间失败 -> " + err.Error())
			os.Exit(1)
		}
	}
}

// func Create(name string) (file *File, err error)
// func OpenFile(name string, flag int, perm FileMode) (file *File, err error)
//这里的flag int是用下面的这些方式，O_CREAT
// const (
//     O_RDONLY int = syscall.O_RDONLY // 只读模式打开文件
//     O_WRONLY int = syscall.O_WRONLY // 只写模式打开文件
//     O_RDWR   int = syscall.O_RDWR   // 读写模式打开文件
//     O_APPEND int = syscall.O_APPEND // 写操作时将数据附加到文件尾部
//     O_CREATE int = syscall.O_CREAT  // 如果不存在将创建一个新文件
//     O_EXCL   int = syscall.O_EXCL   // 和O_CREATE配合使用，文件必须不存在
//     O_SYNC   int = syscall.O_SYNC   // 打开文件用于同步I/O
//     O_TRUNC  int = syscall.O_TRUNC  // 如果可能，打开时清空文件
// )
//打开 以只读,文件不存在将创建 方式打开  要存放的路径资源
// f, f_err := os.OpenFile(file_save, os.O_WRONLY|os.O_CREATE, 0666)
// 		if f_err != nil {
// 					fmt.Fprintf(w, "file open fail:%s", f_err)
// 							}		//文件 copy
// 									_, copy_err := io.Copy(f, file)
// 											if copy_err != nil {
// 														fmt.Fprintf(w, "file copy fail:%s", copy_err)
// 																}		//关闭对应打开的文件		defer f.Close()		defer file.Close()

//获得文件名最快的代码实现方式比较
// func main() {
//     filename := "/root/Desktop/模特唐嘉灵2.jpg"
//     //方案一
//     startTime1 := time.Now()
//     for i := 0; i < 1000000; i++ {
//         _ = filepath.Base(filename)
//     }
//     endTime1 := time.Now()

// 当每次读取块的大小小于4KB，建议使用bufio.NewReader(f), 大于4KB用bufio.NewReaderSize(f,缓存大小)
// 要读Reader, 图方便用ioutil.ReadAll()
// 一次性读取文件，使用ioutil.ReadFile()
// func Ioutil(name string) {
//     if contents,err := ioutil.ReadFile(name);err == nil {
//         //因为contents是[]byte类型，直接转换成string类型后会多一行空格,需要使用strings.Replace替换换行符
//         result := strings.Replace(string(contents),"\n","",1)
//         fmt.Println("Use ioutil.ReadFile to read a file:",result)
//         }
//     }
// func main() {
//     b, err := ioutil.ReadFile("test.log")
//     if err != nil {
//         fmt.Print(err)
//     }
//     fmt.Println(b)
//     str := string(b)
//     fmt.Println(str)
// }
// 读文件方式一：利用ioutil.ReadFile直接从文件读取到[]byte中
// func Read0()  (string){
//     f, err := ioutil.ReadFile("file/test")
//     if err != nil {
//         fmt.Println("read fail", err)
//     }
//     return string(f)
// }

// func read3(path string)string{
//     fi,err := os.Open(path)
//     if err != nil{panic(err)}
//     defer fi.Close()
//     fd,err := ioutil.ReadAll(fi)
//     // fmt.Println(string(fd))
//     return string(fd)
// }
