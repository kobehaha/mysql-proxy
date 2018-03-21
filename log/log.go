package log

import  (
    l4g "github.com/cfxks1989/log4go"
    "path"
    "github.com/kobehaha/mysql-proxy/config"
)

var logger l4g.Logger

func Init(config *config.ServerConfig)  {

    loglevelMap := map[string]l4g.Level{
        "DEBUG":   l4g.DEBUG,
        "TRACE":   l4g.TRACE,
        "INFO":    l4g.INFO,
        "WARNING": l4g.WARNING,
        "ERROR":   l4g.ERROR,
    }
    logFileName := path.Join(config.LogFilePath, "mysql-proxy")
    logger = make(l4g.Logger)
    fileLogWriter := l4g.NewFileLogWriter(logFileName, false)
    fileLogWriter.SetFormat("[%D %T] [%L] (%S) %M")
    fileLogWriter.SetRotate(true)
    fileLogWriter.SetRotateDaily(true)

    if loglevel, found := loglevelMap[config.LogLevel]; found {
        logger.AddFilter("file", loglevel, fileLogWriter)
        if loglevel == l4g.DEBUG {
            logger.AddFilter("file", loglevel, fileLogWriter)
        }
    } else {
        // default is INFO
        logger.AddFilter("file", l4g.INFO, fileLogWriter)
    }

}

func GetLogger()  l4g.Logger {
    return logger
}
