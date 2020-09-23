/**
 * Created by zc on 2020/9/5.
 */
package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zc2638/drone-control/cmd/app"
	"os"
	_ "time/tzdata"
)

func main() {
	command := app.NewServerCommand()
	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
