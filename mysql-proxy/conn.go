package mysql

import (
    "sync"
    "net"
    "encoding/binary"
    "bytes"
    "errors"
    "sync/atomic"
    "runtime"
    "reflect"
    "unsafe"
    "github.com/kobehaha/mysql-proxy/log"
    "github.com/kobehaha/mysql-proxy/constant"
    pki "github.com/kobehaha/mysql-proxy/packetio"
    "github.com/kobehaha/mysql-proxy/util"
    "fmt"
    "strings"
)

var DEFAULT_CAPABILITY uint32 = constant.CLIENT_LONG_PASSWORD | constant.CLIENT_LONG_FLAG |
        constant.CLIENT_CONNECT_WITH_DB | constant.CLIENT_PROTOCOL_41 |
        constant.CLIENT_TRANSACTIONS | constant.CLIENT_SECURE_CONNECTION


type Conn struct {

    sync.Mutex

    c net.Conn

    server *Server

    capability uint32

    connectionId uint32

    status uint16

    collation CollationId

    user string

    db string

    charset   string

    salt []byte

    closed bool

    lastInsertId int64

    affectedRows int64

    pkg *pki.PacketIo

    stmtId uint32




}

var baseConnId uint32 = 20000


func (s *Server) newConn(co net.Conn) *Conn {
    c := new(Conn)

    c.c = co

    c.pkg = pki.NewPacketIO(co)

    c.server = s

    c.c = co
    c.pkg.Sequence = 0

    c.connectionId = atomic.AddUint32(&baseConnId, 1)

    c.status = constant.SERVER_STATUS_AUTOCOMMIT

    c.salt = util.RandomBuf(20)

    c.closed = false

    c.collation = DEFAULT_COLLATION_ID
    c.charset = DEFAULT_CHARSET

    c.stmtId = 0

    return c
}


/******************************************************************************
*                           Initialisation Process                            *
******************************************************************************/

/**
 * From server to client during initial handshake.
 *
 * <pre>
 * Bytes                        Name
 * -----                        ----
 * 1                            protocol_version
 * n (Null-Terminated String)   server_version
 * 4                            thread_id
 * 8                            scramble_buff
 * 1                            (filler) always 0x00
 * 2                            server_capabilities
 * 1                            server_language
 * 2                            server_status
 * 13                           (filler) always 0x00 ...
 * 13                           rest of scramble_buff (4.1)
 * </pre>
 *
 */
// Handshake Initialization Packet
// http://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::Handshake
func (c *Conn) writeInitHandshake() error {

    data := make([]byte, 4, 128)

    // 1 byte 协议版本号

    data = append(data, 10)

    //  n bytes 服务版本号[00]

    data = append(data, constant.ServerVersion...)
    data = append(data, 0)

    // 4 bytes 线程id
    data = append(data, byte(c.connectionId), byte(c.connectionId>>8), byte(c.connectionId>>16), byte(c.connectionId>>24) )

    // 8 bytes 随机数

    data = append(data, c.salt[0:8]...)

    // 1 bytes 填充值 0x00
    data = append(data, 0)


    // 2 bytes 服务器权能标志
    data = append(data, byte(DEFAULT_CAPABILITY), byte(DEFAULT_CAPABILITY>>8))

    // 1 bytes 字符编码
    data = append(data, uint8(DEFAULT_COLLATION_ID))

    // 2 bytes 服务器状态
    data = append(data, byte(c.status), byte(c.status>>8))

    // 2 bytes 服务器权能标志(高16位)
    data = append(data, byte(DEFAULT_CAPABILITY>>16), byte(DEFAULT_COLLATION_ID>>24))

    // 1 bytes 挑战长度
    data = append(data, 0x15)

    // 10 bytes 填充数据
    data = append(data, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)

    // n bytes 挑战随机数
    data = append(data, c.salt[8:]...)

    // 1 byte 结尾0x00
    data = append(data,0)

    return c.pkg.WritePacket(data)

}

func (c *Conn) readHandshakeResponse() error {

    data, err := c.readPacket()
    if err != nil {
        return err
    }

    pos := 0

    // 2 bytes 客户端权能标志 2 bytes 客户端权能标志扩展
    c.capability = binary.LittleEndian.Uint32(data[:4])

    pos = pos + 4

    // 4 bytes 最大消息长度
    pos = pos + 4
    // 1 bytes 字符编码
    pos++

    // 23 bytes 填充数据
    pos += 23

    // n bytes 用户名 Null 结尾
    c.user = string(data[pos: pos + bytes.IndexByte(data[pos:],0)])

    pos += len(c.user) + 1

    // n bytes 认证数据 二进制 Binary 数据
    // n bytes [ 1 为数据长度]

    authLen := int(data[pos])
    pos++

    auth := data[pos:pos+authLen]

    check := util.CheckPassword(c.salt, []byte(c.server.password))

    if !bytes.Equal(auth, check) {
        return errors.New("password is error")
    }

    pos += authLen

    if c.capability&constant.CLIENT_CONNECT_WITH_DB > 0 {

        if len(data[pos:]) == 0 {
            return nil
        }
        // n bytes Null 结尾数据
        db := string(data[pos : pos + bytes.IndexByte(data[pos:], 0)])
        pos += len(c.db) + 1

        if err := c.useDb(db); err != nil {
            return err
        }

    }

    return nil

}

func (c *Conn)useDb(dbname string) error {

    if c.server.database != dbname {
        errors.New("error database not exist")
    }
    c.db = dbname
    return nil
}

func (c *Conn)writePacket(data []byte) error {
    return c.pkg.WritePacket(data)
}

func (c *Conn)readPacket() ([]byte, error) {
    return c.pkg.ReadPacket()
}



func (c *Conn) Handshake() error {
    if err := c.writeInitHandshake(); err != nil {
        log.GetLogger().Error("write Init Handshake error %s", err.Error())
        return err
    }

    log.GetLogger().Debug("init handshak succss")
    if err := c.readHandshakeResponse(); err != nil {
        log.GetLogger().Error("rev handshake response error %s", err.Error())
        return err
    }

    log.GetLogger().Debug("read handshak response")
    if err := c.writeOk(nil); err != nil {
        log.GetLogger().Error("write ok package fiale %s", err.Error())
        return err
    }

    log.GetLogger().Debug("write ok success")
    c.pkg.Sequence = 0
   return nil
}

func (c *Conn) writeOk(r *Result) error {
    if r == nil {
        r = &Result{Status:c.status}
    }

    data := make([]byte , 4, 32)

    data = append(data , constant.OK_HEADER)


    data = append(data, util.PutLengthEncodedInt(r.AffectedRows)...)
    data = append(data, util.PutLengthEncodedInt(r.InsertId)...)


    if c.capability&constant.CLIENT_PROTOCOL_41 > 0 {
        data = append(data, byte(r.Status), byte(r.Status>>8))
        data = append(data, 0, 0)
    }

    return c.writePacket(data)




}


func (c *Conn) Run() {
    defer func() {
        r := recover()
        if err, ok := r.(error); ok {
            const size = 4096
            buf := make([]byte, size)
            buf = buf[:runtime.Stack(buf, false)]

            log.GetLogger().Error("error %v, %s", err, buf)
        }
        c.Close()
    }()

    for {
        data, err := c.readPacket()

        if err != nil {
            return
        }

        if err := c.dispatch(data); err != nil {
            log.GetLogger().Error("dispatch error %s", err.Error())
            //todo 错误数据包
            }
        }

        if c.closed {
            return
        }

        c.pkg.Sequence = 0
    }


func (c *Conn) Close() error {
    if c.closed {
        return nil
    }

    c.c.Close()

    c.closed = true

    return nil
}

func (c *Conn) handlerQuery(sql string) (err error){
    defer func() {
        if e := recover(); e != nil {
            log.GetLogger().Error("execute %s err %v", sql, e)
            return
        }
    }()

    sql = strings.TrimRight(sql, ";")
    fmt.Println("sql is %s", sql)
    return nil

}

func (c *Conn) dispatch(data []byte) error {
    cmd := data[0]
    data = data[1:]

    switch cmd {
    case constant.COM_QUIT:
        c.Close()
        return nil
    case constant.COM_PING:
        return c.writeOk(nil)
    case constant.COM_QUERY:
        //todo 增加query 数据
        c.handlerQuery(String(data))
        return c.writeOk(nil)
    case constant.COM_INIT_DB:
        if err := c.useDb(String(data)); err != nil {
            return err
        } else {
            return c.writeOk(nil)
        }
    default:
        log.GetLogger().Error("command is %s", cmd)
        return errors.New("unknow command not support" + string(cmd))
    }
    return nil
}


// no copy to change slice to string
// use your own risk
func String(b []byte) (s string) {
    pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
    pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
    pstring.Data = pbytes.Data
    pstring.Len = pbytes.Len
    return
}


