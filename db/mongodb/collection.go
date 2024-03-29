package mongodb

import (
	"gopkg.in/mgo.v2"
)

type Collection struct {
	db      *Database
	name    string
	Session *mgo.Collection
}

func (c *Collection) Connect() {
	session := *c.db.session.C(c.name)
	c.Session = &session
}

func NewCollectionSession(name string) *Collection {
	var c = Collection{
		db:   newDBSession(MongoConfig.DBName),
		name: name,
	}
	c.Connect()
	return &c
}
func NewCollection(name string) *mgo.Collection {
	var c = Collection{
		db:   newDBSession(MongoConfig.DBName),
		name: name,
	}
	c.Connect()
	return c.Session
}
func (c *Collection) Close() {
	service.Close(c)
}
