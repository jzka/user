package user

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"github.com/go-kit/kit/log"
	"github.com/user/dbOperations"
)

var (
	ErrUnauthorized = errors.New("Unauthorized")
)

// Service is the user service, providing operations for users to login, register, and retrieve user information.
type Service interface {
	Login(username, password string) (dbOperations.User, error)
	Register(username, password, email, firstname, lastname, phone string) (dbOperations.User, error)
	PostUser(u dbOperations.User) (dbOperations.User, error)
	GetUsers() ([]dbOperations.User, error)
	GetUser(id string) (dbOperations.User, error)
	PostAddress(u dbOperations.Address, userid string) (string, error)
	GetAddresses() ([]dbOperations.Address, error)
	GetAddress(id string) (dbOperations.Address, error)
	DeleteAddress(addrid, userid string) error
	DeleteUser(userid string) error
	//Health() []Health // GET /health
}

type userService struct {
	db     *dbOperations.Mongo
	logger log.Logger
}

func NewUserService(db *dbOperations.Mongo, logger log.Logger) Service {
	return &userService{
		db:     db,
		logger: logger,
	}
}

func (s *userService) Login(username, password string) (dbOperations.User, error) {
	u, err := s.db.GetUserWithName(username)
	if err != nil {
		return u, err
	}
	if u.Password != computeHashFor(password, u.Salt) {
		return u, ErrUnauthorized
	}
	s.db.PopulateAddressesForUser(&u)
	return u, nil
}

func (s *userService) Register(username, password, email, firstname, lastname, phone string) (dbOperations.User, error) {
	u := dbOperations.NewUser()
	u.Email = email
	u.Username = username
	u.FirstName = firstname
	u.LastName = lastname
	u.Phone = phone
	u.Password = computeHashFor(password, u.Salt)
	err := s.db.CreateUser(&u)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (s *userService) PostUser(u dbOperations.User) (dbOperations.User, error) {
	u.NewSalt()
	u.Password = computeHashFor(u.Password, u.Salt)
	err := s.db.CreateUser(&u)
	return u, err
}

func (s *userService) GetUsers() ([]dbOperations.User, error) {
	usrs, err := s.db.GetUsers()
	return usrs, err
}

func (s *userService) GetUser(id string) (dbOperations.User, error) {
	usrs, err := s.db.GetUser(id)
	return usrs, err
}

func (s *userService) PostAddress(addr dbOperations.Address, userid string) (string, error) {
	err := s.db.CreateAddress(&addr, userid)
	return addr.ID, err
}

func (s *userService) GetAddresses() ([]dbOperations.Address, error) {
	addrs, err := s.db.GetAddresses()
	return addrs, err
}

func (s *userService) GetAddress(id string) (dbOperations.Address, error) {
	addr, err := s.db.GetAddress(id)
	return addr, err
}

func (s *userService) DeleteAddress(addrid, userid string) error {
	err := s.db.DeleteAddress(addrid, userid)
	return err
}

func (s *userService) DeleteUser(userid string) error {
	err := s.db.DeleteUser(userid)
	return err
}

func computeHashFor(pass, salt string) string {
	hash := sha256.New()
	io.WriteString(hash, pass)
	io.WriteString(hash, salt)

	return fmt.Sprintf("%x", hash.Sum(nil))
}
