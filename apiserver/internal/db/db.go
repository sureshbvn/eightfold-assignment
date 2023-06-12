// Copyright 2023
//
// Author: Suresh Bysani
//
// This file contains db related struct methods, constructors and utils.
//
// GO-PG is an ORM tool to interact with postgres SQL. It can be thought of as hibernate equivalent.

package db

import (
	"context"
	"fmt"

	"github.com/go-pg/pg/v10"
	"github.com/golang/glog"
	"github.com/spf13/viper"

	"apiserver/internal/config"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

type dbLogger struct{}

func (d dbLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	val, err := q.FormattedQuery()
	if err != nil {
		glog.Errorln(err.Error())
	}
	glog.Infoln(string(val))
	return c, nil
}

func (d dbLogger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	val, err := q.FormattedQuery()
	if err != nil {
		glog.Errorln(err.Error())
	}
	glog.Infoln(string(val))
	return nil
}

// NewDB returns a new instance of go pg DB object. Using this object the postgres queries can be made.
// Please note that this will also connect to the postgres db.
func NewDB(conf *viper.Viper) *pg.DB {
	host := conf.GetString(config.KHost)
	port := conf.GetInt(config.KPort)
	username := conf.GetString(config.KUsername)
	password := conf.GetString(config.KPassword)
	dbname := conf.GetString(config.KDatabaseName)

	// Printing this information to make sure the config is correctly loaded into the config object.
	// TODO(SURESH BYSANI): Move this V2 logging to reduce the logging.
	glog.Infoln("the host", host)
	glog.Infoln("the port", port)
	glog.Infoln("the username", username)
	glog.Infoln("the password", password)
	glog.Infoln("the dbname", dbname)

	db := pg.Connect(&pg.Options{
		User:     username,
		Password: password,
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Database: dbname,
	})

	db.AddQueryHook(dbLogger{})

	return db
}

//----------------------------------------------------------------------------------------------------------------------
