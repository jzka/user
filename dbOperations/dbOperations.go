package dbOperations

import (
	"errors"
	"flag"
	"net/url"
	"time"

	"fmt"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	mname     string
	mpassword string
	mhost     string
	db        = "users"
	//ErrInvalidHexID represents a entity id that is not a valid bson ObjectID
	ErrInvalidHexID = errors.New("Invalid Id Hex")
)

func init() { // os.Getenv("MONGO_USER")
	flag.StringVar(&mname, "mongo-user", "", "Mongo user")
	flag.StringVar(&mpassword, "mongo-password", "", "Mongo password")
	flag.StringVar(&mhost, "mongo-host", "127.0.0.1:27017", "Mongo host")
}

type Mongo struct {
	Session *mgo.Session
}

// Init MongoDB
func (m *Mongo) Init() error {
	//u := getURL()
	var err error
	m.Session, err = mgo.DialWithTimeout("127.0.0.1", time.Duration(5)*time.Second)
	if err != nil {
		fmt.Println("errrr")
		return err
	}
	return m.EnsureIndexes()
}

//CRUD Operations for User

// CreateUser inserts user to MongoDB
func (m *Mongo) CreateUser(u *User) error {
	s := m.Session.Copy()
	defer s.Close()
	id := bson.NewObjectId()
	dbu := NewDBUser()
	dbu.ID = id
	dbu.User = *u
	c := s.DB("").C("users")
	_, err := c.UpsertId(dbu.ID, dbu)
	if err != nil {
		return err
	}
	dbu.User.UserID = dbu.ID.Hex()
	*u = dbu.User
	return nil
}

//GetUser returns user with given id
func (m *Mongo) GetUser(id string) (User, error) {
	s := m.Session.Copy()
	defer s.Close()
	if !bson.IsObjectIdHex(id) {
		return User{}, ErrInvalidHexID
	}
	c := s.DB("").C("users")
	dbu := NewDBUser()
	err := c.FindId(bson.ObjectIdHex(id)).One(&dbu)
	dbu.ConvertObjectsIds()
	if err != nil {
		return User{}, err
	}
	return dbu.User, nil
}

//GetUserWithName returns user with given username
func (m *Mongo) GetUserWithName(username string) (User, error) {
	s := m.Session.Copy()
	defer s.Close()
	c := s.DB("").C("users")
	dbu := NewDBUser()
	err := c.Find(bson.M{"username": username}).One(&dbu)
	dbu.User.UserID = dbu.ID.Hex()
	return dbu.User, err
}

//GetUsers returns all users in db
func (m *Mongo) GetUsers() ([]User, error) {
	s := m.Session.Copy()
	defer s.Close()
	var dbusers []DBUser
	users := make([]User, 0)
	c := s.DB("").C("customers")
	err := c.Find(nil).All(&dbusers)
	if err != nil {
		return users, err
	}
	for _, dbu := range dbusers {
		dbu.ConvertObjectsIds()
		users = append(users, dbu.User)
	}
	return users, nil
}

//PopulateAddressesForUser populates addr fields for given user
func (m *Mongo) PopulateAddressesForUser(u *User) error {
	s := m.Session.Copy()
	defer s.Close()
	ids := make([]bson.ObjectId, 0)
	for _, a := range u.Addresses {
		if !bson.IsObjectIdHex(a.ID) {
			return ErrInvalidHexID
		}
		ids = append(ids, bson.ObjectIdHex(a.ID))
	}
	var dba []DBAddress
	c := s.DB("").C("addresses")
	err := c.Find(bson.M{"_id": bson.M{"$in": ids}}).All(&dba)
	if err != nil {
		return err
	}
	adds := make([]Address, 0)
	for _, a := range dba {
		a.Address.ID = a.ID.Hex()
		adds = append(adds, a.Address)
	}
	u.Addresses = adds
	return nil
}

//DeleteUser deletes user with id
func (m *Mongo) DeleteUser(id string) error {
	if !bson.IsObjectIdHex(id) {
		return ErrInvalidHexID
	}
	s := m.Session.Copy()
	defer s.Close()
	c := s.DB("").C("users")
	u, err := m.GetUser(id)
	if err != nil {
		return err
	}
	addrIds := make([]bson.ObjectId, 0)
	for _, a := range u.Addresses {
		addrIds = append(addrIds, bson.ObjectIdHex(a.ID))
	}
	addrc := s.DB("").C("addresses")
	addrc.RemoveAll(bson.M{"_id": bson.M{"$in": addrIds}})
	errc := c.Remove(bson.M{"_id": bson.ObjectIdHex(id)})
	return errc
}

//CRUD Operations for Addresses

//CreateAddress inserts new address and updates user with addr id
func (m *Mongo) CreateAddress(addr *Address, userId string) error {
	s := m.Session.Copy()
	defer s.Close()
	if !bson.IsObjectIdHex(userId) {
		return ErrInvalidHexID
	}
	c := s.DB("").C("addresses")
	aid := bson.NewObjectId()
	dbAdr := DBAddress{Address: *addr, ID: aid}
	_, err := c.UpsertId(dbAdr.ID, dbAdr)
	if err != nil {
		return err
	}
	errad := m.addIdToUserAddresses(dbAdr.ID, userId)
	dbAdr.Address.ID = dbAdr.ID.Hex()
	*addr = dbAdr.Address
	return errad
}

//GetAddress return an address with given id
func (m *Mongo) GetAddress(id string) (Address, error) {
	s := m.Session.Copy()
	defer s.Close()
	if !bson.IsObjectIdHex(id) {
		return Address{}, ErrInvalidHexID
	}
	c := s.DB("").C("addresses")
	dbAddr := DBAddress{}
	err := c.FindId(bson.ObjectIdHex(id)).One(&dbAddr)
	dbAddr.Address.ID = dbAddr.ID.Hex()
	return dbAddr.Address, err
}

//GetAddresses returns all addresses
func (m *Mongo) GetAddresses() ([]Address, error) {
	s := m.Session.Copy()
	defer s.Close()
	c := s.DB("").C("addresses")
	var dbAdrs []DBAddress
	err := c.Find(nil).All(&dbAdrs)
	adrs := make([]Address, 0)
	for _, dbAdr := range dbAdrs {
		dbAdr.Address.ID = dbAdr.ID.Hex()
		adrs = append(adrs, dbAdr.Address)
	}
	return adrs, err
}

//GetAddressesForUser returns all addresses for a given user
func (m *Mongo) GetAddressesForUser(userid string) ([]Address, error) {
	s := m.Session.Copy()
	defer s.Close()
	adds := make([]Address, 0)
	if !bson.IsObjectIdHex(userid) {
		return adds, ErrInvalidHexID
	}
	c := s.DB("").C("users")
	dbu := NewDBUser()
	errU := c.FindId(bson.ObjectIdHex(userid)).One(&dbu)
	if errU != nil {
		return adds, errU
	}
	var dba []DBAddress
	ac := s.DB("").C("addresses")
	errA := ac.Find(bson.M{"_id": bson.M{"$in": dbu.AddressIDs}}).All(&dba)
	if errA != nil {
		return adds, errA
	}

	for _, a := range dba {
		a.Address.ID = a.ID.Hex()
		adds = append(adds, a.Address)
	}
	return adds, nil
}

//DeleteAddress deletes an address for give userid and addr id
func (m *Mongo) DeleteAddress(userid, addid string) error {
	if !bson.IsObjectIdHex(userid) || !bson.IsObjectIdHex(addid) {
		return ErrInvalidHexID
	}
	s := m.Session.Copy()
	defer s.Close()
	err := m.removeIdFromUserAddresses(bson.ObjectIdHex(addid), userid)
	if err != nil {
		return err
	}
	ac := s.DB("").C("addresses")
	errR := ac.Remove(bson.M{"_id": bson.ObjectIdHex(addid)})
	if errR != nil {
		return errR
	}
	return nil
}

func (m *Mongo) addIdToUserAddresses(id bson.ObjectId, userId string) error {
	s := m.Session.Copy()
	defer s.Close()
	c := s.DB("").C("users")
	return c.Update(bson.M{"_id": bson.ObjectIdHex(userId)},
		bson.M{"$addToSet": bson.M{"addresses": id}})
}

func (m *Mongo) removeIdFromUserAddresses(id bson.ObjectId, userId string) error {
	s := m.Session.Copy()
	defer s.Close()
	c := s.DB("").C("users")
	return c.Update(bson.M{"_id": bson.ObjectIdHex(userId)},
		bson.M{"$pull": bson.M{"addresses": id}})
}

func getURL() url.URL {
	ur := url.URL{
		Scheme: "mongodb",
		Host:   mhost,
		Path:   db,
	}
	if mname != "" {
		u := url.UserPassword(mname, mpassword)
		ur.User = u
	}
	fmt.Println(ur.String)
	return ur
}

// EnsureIndexes ensures username is unique
func (m *Mongo) EnsureIndexes() error {
	s := m.Session.Copy()
	defer s.Close()
	i := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     false,
	}
	c := s.DB("").C("users")
	return c.EnsureIndex(i)
}

//Ping checks db connection
func (m *Mongo) Ping() error {
	s := m.Session.Copy()
	defer s.Close()
	return s.Ping()
}
