package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const FileName = ".nameList"

type PinYin []string

func (s PinYin) Len() int      { return len(s) }
func (s PinYin) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s PinYin) Less(i, j int) bool {
	a, err := UTF82GBK(s[i])
	if err != nil {
		log.Fatalln(err)
	}
	b, err := UTF82GBK(s[j])
	if err != nil {
		log.Fatalln(err)
	}
	bLen := len(b)
	for idx, chr := range a {
		if idx > bLen-1 {
			return false
		}
		if chr != b[idx] {
			return chr < b[idx]
		}
	}
	return true
}

func main() {
	//打开文件
	filePath, err := GetCurrentPath()
	if err != nil {
		log.Fatalln(err)
		return
	}
	filePath += FileName
	log.Println("Open file:", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln(err)
		return
	}

	//按GB18030编码排序
	br := bufio.NewReader(file)
	var nameList []string
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		nameList = append(nameList, string(a))
	}
	sort.Sort(PinYin(nameList))
	if err = file.Close(); err != nil {
		log.Fatalln(err)
	}

	//按顺序重写文件
	file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatalln(err)
		}
	}()
	for _, line := range nameList {
		if _, err = file.WriteString(line + "\n"); err != nil {
			log.Fatalln(err)
			return
		}
		fmt.Println(line)
	}

	//添加新名字
	var str string
	for {
		n, err := fmt.Scanln(&str)
		if n <= 0 || err != nil {
			log.Println("拜拜")
			return
		}
		if isExist(nameList, str) {
			log.Println(str, "已经存在")
			continue
		}
		if _, err = file.WriteString(str + "\n"); err != nil {
			log.Fatalln(err)
			return
		}
	}
}

//获取程序所在路径，而非工作路径
func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return "", errors.New(`can't find "/" or "\"`)
	}
	return string(path[0 : i+1]), nil
}

func isExist(nameList []string, name string) bool {
	for _, n := range nameList {
		if n == name {
			return true
		}
	}
	return false
}

//UTF82GBK : transform UTF8 rune into GBK byte array
func UTF82GBK(src string) ([]byte, error) {
	GB18030 := simplifiedchinese.All[0]
	return ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(src)), GB18030.NewEncoder()))
}

//GBK2UTF8 : transform  GBK byte array into UTF8 string
func GBK2UTF8(src []byte) (string, error) {
	GB18030 := simplifiedchinese.All[0]
	b, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(src), GB18030.NewDecoder()))
	return string(b), err
}
