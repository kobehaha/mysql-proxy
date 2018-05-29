package backend

import (
    "time"
    "net"
    "errors"
    "fmt"
    "bytes"
    "encoding/binary"
    "github.com/kobehaha/mysql-proxy/mysql-proxy"
    pki "github.com/kobehaha/mysql-proxy/packetio"
    constant "github.com/kobehaha/mysql-proxy/constant"
    util "github.com/kobehaha/mysql-proxy/util"
)

var (
    pingPeriod = int64(time.Second*30)
)

type Conn struct {

    conn net.Conn

    pkg *pki.PacketIo

    addr string

    user string

    password string

    db string

    capability uint32

    status uint16

    collation mysql.CollationId

    charset string

    salt []byte

    lastPing int64

    pkgErr error

}

func (c *Conn) Connect(add string, user string, password string, db string) error{

    c.addr = add
    c.user = user
    c.password = password
    c.db = db

    // use utf8
    c.collation = mysql.DEFAULT_COLLATION_ID

    c.charset = mysql.DEFAULT_CHARSET

    return c.ReConnect()

}


func (c *Conn) ReConnect() error {

   if c.conn != nil {
       c.conn.Close()
   }

    n := "tcp"

    netConn, err := net.Dial(n, c.addr)

    if err != nil {
        return err
    }

    c.conn = netConn

    c.pkg = pki.NewPacketIO(netConn)


    if err := c.readInitialHandshake(); err != nil {
        c.conn.Close()
        return err
    }

    if err := c.writeAuthHandshake(); err != nil {
        c.conn.Close()
        return err
    }

    if _, err := c.readOK(); err != nil {
        c.conn.Close()
        return err
    }

    //we must always use autocommit
    if !c.IsAutoCommit() {
        if _, err := c.exec("set autocommit = 1"); err != nil {
            c.conn.Close()

            return err
        }
    }

    c.lastPing = time.Now().Unix()

    return nil


}


func (c *Conn) readOK() (*mysql.Result, error) {
    data, err := c.readPacket()
    if err != nil {
        return nil, err
    }

    if data[0] == constant.OK_HEADER {
        return c.handleOKPacket(data)
    } else if data[0] == constant.ERR_HEADER {
        return nil, nil
    } else {
        return nil, errors.New("invalid ok packet")
    }
}



func (c *Conn) handleOKPacket(data []byte) (*mysql.Result, error) {
    var n int
    var pos int = 1

    r := new(mysql.Result)

    r.AffectedRows, _, n = util.LengthEncodedInt(data[pos:])
    pos += n
    r.InsertId, _, n = util.LengthEncodedInt(data[pos:])
    pos += n

    if c.capability&constant.CLIENT_PROTOCOL_41 > 0 {
        r.Status = binary.LittleEndian.Uint16(data[pos:])
        c.status = r.Status
        pos += 2

        //todo:strict_mode, check warnings as error
        //Warnings := binary.LittleEndian.Uint16(data[pos:])
        //pos += 2
    } else if c.capability&constant.CLIENT_TRANSACTIONS > 0 {
        r.Status = binary.LittleEndian.Uint16(data[pos:])
        c.status = r.Status
        pos += 2
    }

    //info
    return r, nil
}




func (c *Conn) exec(query string) (*mysql.Result, error) {

    if err := c.writeCommandStr(constant.COM_QUERY, query); err != nil {
        return nil, err
    }
    return nil,nil

    //return c.readResult(false)
}





func (c *Conn) writeCommandStr(command byte, arg string) error {
    c.pkg.Sequence = 0

    length := len(arg) + 1

    data := make([]byte, length+4)

    data[4] = command

    copy(data[5:], arg)

    return c.writePacket(data)
}





func (c *Conn) Close() error {
    if c.conn != nil {
        c.conn.Close()
        c.conn = nil
    }

    return nil
}

func (c *Conn) readPacket() ([]byte, error) {
    d, err := c.pkg.ReadPacket()
    c.pkgErr = err
    return d, err
}

func (c *Conn) writePacket(data []byte) error {
    err := c.pkg.WritePacket(data)
    c.pkgErr = err
    return err
}


func (c *Conn) readInitialHandshake() error {
    data, err := c.readPacket()
    if err != nil {
        return err
    }

    if data[0] == constant.ERR_HEADER {
        return errors.New("read initial handshake error")
    }

    if data[0] < constant.MinProtocolVersion {
        return fmt.Errorf("invalid protocol version %d, must >= 10", data[0])
    }

    //skip mysql version and connection id
    //mysql version end with 0x00
    //connection id length is 4
    pos := 1 + bytes.IndexByte(data[1:], 0x00) + 1 + 4

    c.salt = append(c.salt, data[pos:pos+8]...)

    //skip filter
    pos += 8 + 1

    //capability lower 2 bytes
    c.capability = uint32(binary.LittleEndian.Uint16(data[pos : pos+2]))

    pos += 2

    if len(data) > pos {
        //skip server charset
        //c.charset = data[pos]
        pos += 1

        c.status = binary.LittleEndian.Uint16(data[pos : pos+2])
        pos += 2

        c.capability = uint32(binary.LittleEndian.Uint16(data[pos:pos+2]))<<16 | c.capability

        pos += 2

        //skip auth data len or [00]
        //skip reserved (all [00])
        pos += 10 + 1

        // The documentation is ambiguous about the length.
        // The official Python library uses the fixed length 12
        // mysql-proxy also use 12
        // which is not documented but seems to work.
        c.salt = append(c.salt, data[pos:pos+12]...)
    }

    return nil
}


func (c *Conn) writeAuthHandshake() error {

    capability := constant.CLIENT_PROTOCOL_41 | constant.CLIENT_SECURE_CONNECTION |
            constant.CLIENT_LONG_PASSWORD | constant.CLIENT_TRANSACTIONS | constant.CLIENT_LONG_FLAG

    capability &= c.capability


    //packet length
    //capbility 4
    //max-packet size 4
    //charset 1
    //reserved all[0] 23
    length := 4 + 4 + 1 + 23

    //username
    length += len(c.user) + 1

    //we only support secure connection
    auth := util.CheckPassword(c.salt, []byte(c.password))

    length += 1 + len(auth)

    if len(c.db) > 0 {
        capability |= constant.CLIENT_CONNECT_WITH_DB

        length += len(c.db) + 1
    }

    c.capability = capability

    data := make([]byte, length+4)

    //capability [32 bit]
    data[4] = byte(capability)
    data[5] = byte(capability >> 8)
    data[6] = byte(capability >> 16)
    data[7] = byte(capability >> 24)

    //MaxPacketSize [32 bit] (none)
    //data[8] = 0x00
    //data[9] = 0x00
    //data[10] = 0x00
    //data[11] = 0x00

    //Charset [1 byte]
    data[12] = byte(c.collation)

    //Filler [23 bytes] (all 0x00)
    pos := 13 + 23

    //User [null terminated string]
    if len(c.user) > 0 {
        pos += copy(data[pos:], c.user)
    }
    //data[pos] = 0x00
    pos++

    // auth [length encoded integer]
    data[pos] = byte(len(auth))
    pos += 1 + copy(data[pos+1:], auth)

    // db [null terminated string]
    if len(c.db) > 0 {
        pos += copy(data[pos:], c.db)
        //data[pos] = 0x00
    }

    return c.writePacket(data)


}



func (c *Conn) IsAutoCommit() bool {
    return c.status&constant.SERVER_STATUS_AUTOCOMMIT > 0
}


