package main

import (
	"log"
	"os"
	"strconv"
)

type Env struct {
	s Storage
}

func GetEnv() *Env {
	addr := os.Getenv("APP_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	passwd := os.Getenv("APP_REDIS_PASSWD")
	if passwd == "" {
		passwd = ""
	}

	dbS := os.Getenv("APP_REDIS_DB")
	if dbS == "" {
		dbS = "0"
	}

	db, err := strconv.Atoi(dbS)
	if err != nil {
		panic(err)
	}
	log.Printf("conntect to redis (addr: %s password: %s db: %d)\n", addr, passwd, db)
	return &Env{s: NewRedisClient(addr, passwd, db)}
}
