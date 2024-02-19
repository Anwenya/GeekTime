package main

import (
	"context"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"log"
	"time"
)

func main() {
	cache, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(10*time.Millisecond))

	err := cache.Set("my-unique-key", []byte("value"))
	if err != nil {
		log.Printf("err:%v", err)
		return
	}
	time.Sleep(time.Second * 2)
	entry, error := cache.Get("my-unique-key")
	fmt.Println(error)
	fmt.Println(string(entry))
}
