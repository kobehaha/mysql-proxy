package mysql


import (
    "sync"
    "time"
    "github.com/kobehaha/mysql-proxy/config"
)


const (
    Master = "master"
    Slave  = "slave"
)

type Node struct {

    sync.Mutex

    server *Server

    cfg config.NodeConfig

    downAfterNoAlive time.Duration

    lastMasterPing int64
    lastSlavePing  int64
}
