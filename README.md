#### 简介
flag库基于go标准库修改而来，最近经常写些命令行工具，发现flag库不能满足我的需求，
扩展了下，此项目存放这些私有修改
#### 演示
![flag](./flag.gif)
#### 功能
* flag库的所有功能
* 在所有的flag.type()基础上增加flag.typeSlice函数(其中的type可以是String, int等)
* 同一个命令行选项可以有多个别名(flag.String("opt1, opt2", "", "example opotion"), opt1,opt2代表同一个意思)

#### example
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

运行demo code
* 注意看-H 和 --header都被收集到h slice里面
```shell
env GOPATH=`pwd` go run main.go -H "appkey:123" -H "User-Agent: main" --header "Accept: */*" -f file -file file2
```
* 输出结果
```shell
output-->  h/header([]string{"appkey:123", "User-Agent: main", "Accept: */*"}), f/file("file2"), op()
```

* help 输出效果
```shell
env GOPATH=`pwd` go run main.go -h
Usage of /tmp/go-build520917535/command-line-arguments/_obj/exe/main:
  -H, --header string[]
    	http header
  -dir string
    	open dir
  -f, --file string
    	open audio file
```


