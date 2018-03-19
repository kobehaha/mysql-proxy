package mysql

import (
    "sync"
    "net"
)

var DEFAULT_CAPABILITY uint32 = CLIENT_LONG_PASSWORD | CLIENT_LONG_FLAG |
        CLIENT_CONNECT_WITH_DB | CLIENT_PROTOCOL_41 |
        CLIENT_TRANSACTIONS | CLIENT_SECURE_CONNECTION


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

    salt []byte

    closed bool

    lastInsertId int64

    affectedRows int64

    pkg *PacketIo


}

var baseConnId uint32 = 20000


func (c *Conn) writeInitHandshake() error {

    data := make([]byte, 4, 128)

    // 1 byte 协议版本号

    data = append(data, 10)

    //  n bytes 服务版本号[00]

    data = append(data, ServerVersion...)
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

func (c *Conn)writePacket() ([]byte, error) {
    return c.pkg.ReadPacket()
}

func (c *Conn)readPacket() ([]byte, error) {
    return c.pkg.ReadPacket()
}



func (c *Conn) Handshake() error {
   return nil
}





