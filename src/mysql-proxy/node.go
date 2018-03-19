package mysql


import (
    "sync"
    "time"
    "../config"
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
