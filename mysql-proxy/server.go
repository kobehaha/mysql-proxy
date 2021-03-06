package mysql

import  (
    "net"
    "os"
    "os/signal"
    "syscall"
    "runtime"
    "github.com/kobehaha/mysql-proxy/log"
    "github.com/kobehaha/mysql-proxy/config"
    "fmt"
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

    if server == nil {
        fmt.Println("server is null")
    }
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

    conn := server.newConn(con)

    defer func() {
        if err := recover(); err != nil {
            const size = 4096
            buf := make([]byte, size)
            buf = buf[:runtime.Stack(buf, false)]
            log.GetLogger().Error("onConn panic %v: %v: %s", con.RemoteAddr().String(), err, buf)
        }

        con.Close()
    }()



    if err := conn.Handshake(); err != nil {
        log.GetLogger().Error("handshake error %s", err)
        conn.Close()
        return
    }

    conn.Run()


    //for {
    //    buf := make([]byte, 512)
    //    n, err := con.Read(buf)
    //    if err != nil {
    //       log.GetLogger().Error("read data from socke error %s", err)
    //    }
    //    if n == 0 {
    //        break
    //    }
    //    log.GetLogger().Info("get byte data is %s", string(buf))
    //}
    defer con.Close()

}
