package mysql

import  (
    "net"
    "../config"
    "../log"
    "os"
    "os/signal"
    "syscall"
)

type Server struct {

    cfg *config.ServerConfig

    addr     string
    user     string
    password string
    database string

    running bool

    listener net.Listener

    nodes []string
}

func init(){

}

func NewServer(cfg *config.ServerConfig) (*Server, error) {

    var server = new(Server)
    server.addr = cfg.ListenerAddress
    server.user = cfg.User
    server.database = cfg.Database
    server.password = cfg.Password
    server.nodes = cfg.RemoteNodes

    var err error
    netProto := "tcp"

    server.listener, err = net.Listen(netProto, server.addr)

    if err != nil {
        return nil, err
    }
    log.GetLogger().Info("Server mysql proxy is start and ready Listen  proto is %s and address is %s", netProto, server.addr)


    sc := make(chan os.Signal, 1)
    signal.Notify(sc,
        syscall.SIGHUP,
        syscall.SIGINT,
        syscall.SIGTERM,
        syscall.SIGQUIT)

    go func() {
        sig := <-sc
        log.GetLogger().Info("Got signal [%d] to exit.", sig)
        server.Close()
    }()


    return server, nil
}


func (server *Server) Start() {
    server.running = true

    for server.running {
        conn, err := server.listener.Accept()
        if err != nil {
            log.GetLogger().Error("accept error %s", err.Error())
            continue
        }
        go server.Conn(conn)
    }
}

func (server *Server) Close() {
    server.running = false
    if server.listener != nil {
        server.listener.Close()
    }

}

func (server *Server) Conn(con net.Conn)  {

    for {
        buf := make([]byte, 512)
        n, err := con.Read(buf)
        if err != nil {
           log.GetLogger().Error("read data from socke error %s", err)
        }
        if n == 0 {
            break
        }
        log.GetLogger().Info("get byte data is %s", string(buf))
    }
    defer con.Close()

}
