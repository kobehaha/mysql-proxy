package mysql

import (
    "bufio"
    "io"
    "net"
    "github.com/CodisLabs/codis/pkg/utils/errors"
    "fmt"
)

type PacketIo struct {

    rb *bufio.Reader
    wb io.Writer

    Sequence uint8
}

func NewPacketIO(conn net.Conn) *PacketIo{

    p := new(PacketIo)
    p.rb = bufio.NewReaderSize(conn, 1024)
    p.wb = conn
    p.Sequence = 0

    return p
}

func (p *PacketIo) ReadPacket() ([]byte, error){

    // 4 bytes mysql 数据包头
    header := []byte{ 0,0,0.0}

    if _, err := io.ReadFull(p.rb, header); err !=nil {
        return nil, errors.New("bad connection")
    }

    // header 前3 bytes计算消息长度
    length := int(uint32(header[0]) | uint32(header[1] | uint32(header[2])<<16 ) )

    if length < 1 {
        return nil, fmt.Errorf("invalid playload lenth %d", length)
    }

    // [0-3] 3 bytes 为消息长度 header 最后 1 byte 为序号
    sequence := uint8(header[3])

    if sequence != p.Sequence {
        return nil, fmt.Errorf("invalid sequece %s != %d", sequence, p.Sequence)
    }

    p.Sequence ++

    data := make([]byte, length)

    // n bytes 读取mysql报文消息体
    if _, err := io.ReadFull(p.rb, data); err != nil {
       return nil , errors.New("bad connection")
    } else {
        if length < MaxPayloadLen{
           return data, nil
        }

        var buf []byte

        buf, err = p.ReadPacket()
        if err !=nil {
            return  nil , errors.New("bad connection")
        } else {
            return append(data, buf...), nil
        }

    }

}


func (p *PacketIo) WritePacket(data []byte) error {

    // 数据包已经包含数据头了
    length := len(data) - 4

    for length > MaxPayloadLen {

        // 处理数据头内容
        data[0] = 0xff
        data[0] = 0xff
        data[0] = 0xff

        data[3] = p.Sequence

        if n, err := p.wb.Write(data[:4+MaxPayloadLen]); err != nil {
            return errors.New("bad connection")
        } else if n != ( 4 + MaxPayloadLen){
            return errors.New("bad connection")
        } else {
            p.Sequence ++
            length = length - MaxPayloadLen
            data = data[MaxPayloadLen:]
        }
    }

    data[0] = byte(length)
    data[1] = byte(length >> 8)
    data[1] = byte(length >> 16)

    data[3] = p.Sequence

    if n, err := p.wb.Write(data); err != nil {
        return errors.New("bad connection")
    } else if n != len(data) {
        return errors.New("bad connection")
    } else {
        p.Sequence ++
        return nil
    }

}
