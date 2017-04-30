package dbOperations

import (
	"os"
	"testing"

	"gopkg.in/mgo.v2/dbtest"
)

var (
	TestMongo  = Mongo{}
	TestServer = dbtest.DBServer{}
	TestUser   = User{
		FirstName: "firstname",
		LastName:  "lastname",
		Username:  "username",
		Password:  "password",
		Addresses: []Address{
			Address{
				Street: "street",
			},
		},
	}
	TestAddress = Address{
		Country: "country",
		City:    "city",
	}
)

func init() {
	TestServer.SetPath("/tmp")
}

func TestMain(m *testing.M) {
	TestMongo.Session = TestServer.Session()
	TestMongo.EnsureIndexes()
	TestMongo.Session.Close()
	exitTest(m.Run())
}

func exitTest(i int) {
	TestServer.Wipe()
	TestServer.Stop()
	os.Exit(i)
}

func TestCreateUser(t *testing.T) {
	TestMongo.Session = TestServer.Session()
	defer TestMongo.Session.Close()
	err := TestMongo.CreateUser(&TestUser)
	if err != nil {
		t.Error(err)
	}
	err = TestMongo.CreateUser(&TestUser)
	if err == nil {
		t.Error("expected duplicate key error")
	}
}

func TestGetUser(t *testing.T) {
	TestMongo.Session = TestServer.Session()
	defer TestMongo.Session.Close()
	_, err := TestMongo.GetUser(TestUser.UserID)
	if err != nil {
		t.Error(err)
	}
}

func TestGetUserWithName(t *testing.T) {
	TestMongo.Session = TestServer.Session()
	defer TestMongo.Session.Close()
	_, err := TestMongo.GetUserWithName(TestUser.Username)
	if err != nil {
		t.Error(err)
	}
}

func TestGetUsers(t *testing.T) {
	TestMongo.Session = TestServer.Session()
	defer TestMongo.Session.Close()
	_, err := TestMongo.GetUsers()
	if err != nil {
		t.Error(err)
	}
}

func TestCreateAddress(t *testing.T) {
	TestMongo.Session = TestServer.Session()
	defer TestMongo.Session.Close()
	err := TestMongo.CreateAddress(&TestAddress, TestUser.UserID)
	if err != nil {
		t.Error(err)
	}
}

func TestPopulateAddressesForUser(t *testing.T) {
	TestMongo.Session = TestServer.Session()
	defer TestMongo.Session.Close()
	tu, err := TestMongo.GetUser(TestUser.UserID)
	errd := TestMongo.PopulateAddressesForUser(&tu)
	TestUser = tu
	if errd != nil {
		t.Error(err)
	}
}

func TestGetAddress(t *testing.T) {
	TestMongo.Session = TestServer.Session()
	defer TestMongo.Session.Close()
	tu, err := TestMongo.GetUser(TestUser.UserID)
	if err != nil {
		t.Error(err)
	}
	_, erra := TestMongo.GetAddress(tu.Addresses[0].ID)
	if erra != nil {
		t.Error(erra)
	}
}

func TestGetAddresses(t *testing.T) {
	TestMongo.Session = TestServer.Session()
	defer TestMongo.Session.Close()
	addrs, err := TestMongo.GetAddresses()
	if err != nil {
		t.Error(err)
	}
	if len(addrs) == 0 {
		t.Error("expected addresses")
	}
}

func TestGetAddressesForUser(t *testing.T) {
	TestMongo.Session = TestServer.Session()
	defer TestMongo.Session.Close()
	addrs, err := TestMongo.GetAddressesForUser(TestUser.UserID)
	if err != nil {
		t.Error(err)
	}
	if len(addrs) == 0 {
		t.Error("expected addresses")
	}
}

func TestDeleteAddress(t *testing.T) {
	TestMongo.Session = TestServer.Session()
	defer TestMongo.Session.Close()
	if len(TestUser.Addresses) == 0 {
		t.Error("cant delete something that doesnt exist")
	}
	err := TestMongo.DeleteAddress(TestUser.UserID, TestUser.Addresses[0].ID)
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteUser(t *testing.T) {
	TestMongo.Session = TestServer.Session()
	defer TestMongo.Session.Close()
	err := TestMongo.DeleteUser(TestUser.UserID)
	if err != nil {
		t.Error(err)
	}
}

func TestPing(t *testing.T) {
	TestMongo.Session = TestServer.Session()
	defer TestMongo.Session.Close()
	err := TestMongo.Ping()
	if err != nil {
		t.Error(err)
	}
}
