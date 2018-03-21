package main

import (
    "runtime"
    "os"
    "fmt"
    "github.com/kobehaha/mysql-proxy/config"
    "github.com/kobehaha/mysql-proxy/mysql-proxy"
    "github.com/kobehaha/mysql-proxy/log"
)

var VERSION  = 1.0
var buildstamp = "no timestamp set"

func main() {

    runtime.GOMAXPROCS(runtime.NumCPU())
    var configFile string
    if len(os.Args) == 2 {
        configFile = os.Args[1]
    } else{
        fmt.Printf("CMD eg: mysql-proxy config.json, version: %s, buildstamp: %s\n", VERSION, buildstamp)
        os.Exit(1)
    }
    // ParseConfig
    cfg, err := config.ParseConfigFile(configFile)
    if err != nil {
        fmt.Println("parse config file error")
    }
    log.Init(cfg)
    server, err := mysql.NewServer(cfg)
    if err != nil {
       log.GetLogger().Error("server start error %s", err)
    }
    server.Start()
}





