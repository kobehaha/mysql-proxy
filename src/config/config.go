package config

import (
    "io/ioutil"
    simplejson "github.com/bitly/go-simplejson"
)

type NodeConfig struct {
}

type ServerConfig struct {

    ListenerAddress string
    User string
    Password string
    Database string
    RemoteNodes []string
    LogFilePath     string
    LogLevel    string
}

func ParseConfigFile(file string) (*ServerConfig, error) {
    data, err := ioutil.ReadFile(file)
    if err != nil {
        return nil ,err
    }
    return ParseConfigData(data)

}

func ParseConfigData(data []byte) (*ServerConfig,error ) {
    var config = new(ServerConfig)
    configJson, err := simplejson.NewJson(data)
    if err != nil {
        return nil , err
    }
    config.ListenerAddress = configJson.Get("ListenerAddress").MustString()
    config.RemoteNodes = configJson.Get("RemoteNodes").MustStringArray()
    config.User = configJson.Get("User").MustString()
    config.Password = configJson.Get("Password").MustString()
    config.Database = configJson.Get("Database").MustString()
    config.LogFilePath = configJson.Get("LogfilePath").MustString()
    config.LogLevel = configJson.Get("LogLevel").MustString()

    return config, nil
}