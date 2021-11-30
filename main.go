package main

import (
	"github.com/monostylegc/BabyDoge/cli"
	"github.com/monostylegc/BabyDoge/db"
)

func main() {
	//main함수가 종료될 때 db를 닫는다.
	defer db.Close()
	cli.Start()
}
