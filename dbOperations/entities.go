package dbOperations

import (
	"crypto/sha256"
	"fmt"
	"io"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// User describes specific user fields
type User struct {
	UserID    string    `json:"id" bson:"-"`
	Email     string    `json:"-" bson:"email"`
	Username  string    `json:"username" bson:"username"`
	Password  string    `json:"-" bson:"password,omitempty"`
	FirstName string    `json:"firstName" bson:"firstname"`
	LastName  string    `json:"lastName" bson:"lastname"`
	Phone     string    `json:"phone" bson:"phone"`
	Addresses []Address `json:"-,omitempty" bson:"-"`
	Salt      string    `json:"-" bson:"salt"`
}

// NewUser returns a new user
func NewUser() User {
	u := User{Addresses: make([]Address, 0)}
	u.NewSalt()
	return u
}

//NewSalt creates a hash from current time
func (u *User) NewSalt() {
	h := sha256.New()
	io.WriteString(h, strconv.Itoa(int(time.Now().UnixNano())))
	u.Salt = fmt.Sprintf("%x", h.Sum(nil))
}

// Address describes specific address fields
type Address struct {
	ID        string `json:"id" bson:"-"`
	Country   string `json:"country" bson:"country,omitempty"`
	City      string `json:"city" bson:"city,omitempty"`
	Street    string `json:"street" bson:"street,omitempty"`
	Number    string `json:"number" bson:"number,omitempty"`
	PostCode  string `json:"postcode" bson:"postcode,omitempty"`
	ExtraInfo string `json:"extraInfo" bson:"extraInfo,omitempty"`
}

// DBAddress is a wrapper for Address
type DBAddress struct {
	Address `bson:",inline"`
	ID      bson.ObjectId `bson:"_id"`
}

// DBUser contains User field and bson fields specific to mongoDb
type DBUser struct {
	User       `bson:",inline"`
	ID         bson.ObjectId   `bson:"_id"`
	AddressIDs []bson.ObjectId `bson:"addresses"`
}

//NewDBUser returns a new DBUser
func NewDBUser() DBUser {
	u := NewUser()
	return DBUser{
		User:       u,
		AddressIDs: make([]bson.ObjectId, 0),
	}
}

func (dbu *DBUser) ConvertObjectsIds() {
	dbu.User.UserID = dbu.ID.Hex()
	for _, id := range dbu.AddressIDs {
		dbu.User.Addresses = append(dbu.User.Addresses, Address{ID: id.Hex()})
	}
}
