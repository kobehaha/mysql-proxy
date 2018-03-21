package mysql



const (

    OK_HEADER          byte = 0x00
    ERR_HEADER         byte = 0xff
    EOF_HEADER         byte = 0xfe
    LocalInFile_HEADER byte = 0xfb

    ServerVersion      string = "5.5.31-mysql-proxy-0.1"

    MaxPayloadLen      int    = 1<<24 - 1

)


const (
    CLIENT_LONG_PASSWORD uint32 = 1 << iota
    CLIENT_FOUND_ROWS
    CLIENT_LONG_FLAG
    CLIENT_CONNECT_WITH_DB
    CLIENT_NO_SCHEMA
    CLIENT_COMPRESS
    CLIENT_ODBC
    CLIENT_LOCAL_FILES
    CLIENT_IGNORE_SPACE
    CLIENT_PROTOCOL_41
    CLIENT_INTERACTIVE
    CLIENT_SSL
    CLIENT_IGNORE_SIGPIPE
    CLIENT_TRANSACTIONS
    CLIENT_RESERVED
    CLIENT_SECURE_CONNECTION
    CLIENT_MULTI_STATEMENTS
    CLIENT_MULTI_RESULTS
    CLIENT_PS_MULTI_RESULTS
    CLIENT_PLUGIN_AUTH
    CLIENT_CONNECT_ATTRS
    CLIENT_PLUGIN_AUTH_LENENC_CLIENT_DATA
)


const (
    SERVER_STATUS_IN_TRANS             uint16 = 0x0001
    SERVER_STATUS_AUTOCOMMIT           uint16 = 0x0002
    SERVER_MORE_RESULTS_EXISTS         uint16 = 0x0008
    SERVER_STATUS_NO_GOOD_INDEX_USED   uint16 = 0x0010
    SERVER_STATUS_NO_INDEX_USED        uint16 = 0x0020
    SERVER_STATUS_CURSOR_EXISTS        uint16 = 0x0040
    SERVER_STATUS_LAST_ROW_SEND        uint16 = 0x0080
    SERVER_STATUS_DB_DROPPED           uint16 = 0x0100
    SERVER_STATUS_NO_BACKSLASH_ESCAPED uint16 = 0x0200
    SERVER_STATUS_METADATA_CHANGED     uint16 = 0x0400
    SERVER_QUERY_WAS_SLOW              uint16 = 0x0800
    SERVER_PS_OUT_PARAMS               uint16 = 0x1000
)


const (
    COM_SLEEP byte = iota
    COM_QUIT
    COM_INIT_DB
    COM_QUERY
    COM_FIELD_LIST
    COM_CREATE_DB
    COM_DROP_DB
    COM_REFRESH
    COM_SHUTDOWN
    COM_STATISTICS
    COM_PROCESS_INFO
    COM_CONNECT
    COM_PROCESS_KILL
    COM_DEBUG
    COM_PING
    COM_TIME
    COM_DELAYED_INSERT
    COM_CHANGE_USER
    COM_BINLOG_DUMP
    COM_TABLE_DUMP
    COM_CONNECT_OUT
    COM_REGISTER_SLAVE
    COM_STMT_PREPARE
    COM_STMT_EXECUTE
    COM_STMT_SEND_LONG_DATA
    COM_STMT_CLOSE
    COM_STMT_RESET
    COM_SET_OPTION
    COM_STMT_FETCH
    COM_DAEMON
    COM_BINLOG_DUMP_GTID
    COM_RESET_CONNECTION
)