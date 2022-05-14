package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func RandString(len int) []byte {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return bytes
}


func randombytes()[]byte{
	s1 := make([]byte, 0)
	buf := bytes.NewBuffer(s1)
	var i1 int64 = int64(rand.Intn(3000000))
	buf.Reset()
	binary.Write(buf, binary.LittleEndian, i1)
	return buf.Bytes()

}

var wg sync.WaitGroup

func hello (i int){
	defer wg.Done()	// goroutine结束就登记-1
	fmt.Println("hello Goroutine!",i)
}
func a() (i int) {
	i = 0
	defer  func() {
		fmt.Println(i)
	}() //在有具名返回值的函数中（这里具名返回值为 i），执行 return 2 的时候实际上已经将 i 的值重新赋值为 2。
	i++
	defer fmt.Println(i) //输出1，因为i此时就是1
	return	2
}
func main(){
/*	for i:=0;i<10;i++{
		wg.Add(1)
		go hello(i)
	}
	wg.Wait()*/
	var []string stt=

}