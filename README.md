#### 简介
flag库基于go标准库修改而来，最近经常写些命令行工具，发现flag库不能满足需求，
扩展了下，此项目存放这些私有修改
#### 演示
![flag](./flag.gif)
#### 功能
* flag库的所有功能
* 类似于curl -H 功能
* 命令别名, 有些命令比如curl -F 和curl --form表示同一含义，这时候可以使用命令别名功能
* 子母命令, git 命令是典型的子母命令应用场景，git add 或 git rm

#### curl -H 功能和命令别名示例
```golang
package main

import (
    //"./flag"
    "fmt"
    "github.com/guonaihong/flag"
)

func main() {
    f := flag.String("f,file", "", "open audio file")
    op := flag.String("dir", "", "open dir")
    h := flag.StringSlice("H, header", []string{}, "http header")
    flag.Parse()
    fmt.Printf("output-->  h/header(%#v), f/file(%#v), op(%s)\n", *h, *f, *op)
}

```

* 注意看-H 和 --header都被收集到go里面slice类型变量里面(下面是运行demo code的命令)
```console
env GOPATH=`pwd` go run main.go -H "appkey:123" -H "User-Agent: main" --header "Accept: */*" -f file -file file2
```
* 输出结果
```console
output-->  h/header([]string{"appkey:123", "User-Agent: main", "Accept: */*"}), f/file("file2"), op()
```

* help 输出效果
```console
env GOPATH=`pwd` go run main.go -h
Usage of /tmp/go-build520917535/command-line-arguments/_obj/exe/main:
  -H, --header string[]
    	http header
  -dir string
    	open dir
  -f, --file string
    	open audio file
```

#### 子母命令示示例
如果要实现一个主命令包含3个子命令(http, websocket, tcp)，其中http子命令又包含5个命令行选项，可以使用如下用法
``` golang
func main() {
    parent := flag.NewParentCommand(os.Args[0])

    parent.SubCommand("http", "Use the http subcommand", func() {
        argv0 := os.Args[0]
        argv  := parent.Args()

        commandlLine := flag.NewFlagSet(argv0, flag.ExitOnError)

        headers := commandlLine.StringSlice("H, header", []string{}, "Pass custom header LINE to server (H)")
        forms := commandlLine.StringSlice("F, form", []string{}, "Specify HTTP multipart POST data (H)")
        formStrings := commandlLine.StringSlice("form-string", []string{}, "Specify HTTP multipart POST data (H)")
        URL := commandlLine.String("url", "", "Specify a URL to fetch")
        data := commandlLine.String("d, data", "", "HTTP POST data")

        commandlLine.Author("guonaihong https://github.com/guonaihong/gurl")
        commandlLine.Parse(argv)
    })

    parent.SubCommand("ws, websocket", "Use the websocket subcommand", func() {
        //wsurl.Main(os.Args[0], parent.Args())
    })

    parent.SubCommand("tcp, udp", "Use the tcp or udp subcommand", func() {
        //conn.Main(os.Args[0], parent.Args())
    })

    parent.Parse(os.Args[1:])
}
```
