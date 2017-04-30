package user

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/user/dbOperations"
)

// Endpoints collects the endpoints that comprise the Service.
type Endpoints struct {
	LoginEndpoint       endpoint.Endpoint
	RegisterEndpoint    endpoint.Endpoint
	UserGetEndpoint     endpoint.Endpoint
	UserPostEndpoint    endpoint.Endpoint
	AddressGetEndpoint  endpoint.Endpoint
	AddressPostEndpoint endpoint.Endpoint
	DeleteEndpoint      endpoint.Endpoint
}

// MakeEndpoints returns an Endpoints structure, where each endpoint is
// backed by the given service.
func MakeEndpoints(s Service) Endpoints {
	return Endpoints{
		LoginEndpoint:       MakeLoginEndpoint(s),
		RegisterEndpoint:    MakeRegisterEndpoint(s),
		UserGetEndpoint:     MakeUserGetEndpoint(s),
		UserPostEndpoint:    MakeUserPostEndpoint(s),
		AddressGetEndpoint:  MakeAddressGetEndpoint(s),
		AddressPostEndpoint: MakeAddressPostEndpoint(s),
		DeleteEndpoint:      MakeDeleteEndpoint(s),
	}
}

// MakeLoginEndpoint returns an endpoint via the given service.
func MakeLoginEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(loginRequest)
		u, err := s.Login(req.Username, req.Password)
		return userResponse{User: u}, err
	}
}

// MakeRegisterEndpoint returns an endpoint via the given service.
func MakeRegisterEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(registerRequest)
		usr, err := s.Register(req.Username, req.Password, req.Email, req.FirstName, req.LastName, req.Phone)
		return postResponse{ID: usr.UserID}, err
	}
}

// MakeUserGetEndpoint returns an endpoint via the given service.
func MakeUserGetEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetRequest)
		if req.ID == "" {
			if req.Attr == "" {
				usrs, err := s.GetUsers()
				return EmbedStruct{usersResponse{Users: usrs}}, err
			}
		}
		usr, err := s.GetUser(req.ID)
		if req.Attr == "addresses" {
			return EmbedStruct{addressesResponse{Addresses: usr.Addresses}}, err
		}
		return usr, err
	}
}

// MakeUserPostEndpoint returns an endpoint via the given service.
func MakeUserPostEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(dbOperations.User)
		usr, err := s.PostUser(req)
		return postResponse{ID: usr.UserID}, err
	}
}

// MakeAddressGetEndpoint returns an endpoint via the given service.
func MakeAddressGetEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetRequest)
		if req.ID == "" {
			adds, err := s.GetAddresses()
			return EmbedStruct{addressesResponse{Addresses: adds}}, err
		}
		add, err := s.GetAddress(req.ID)
		return add, err
	}
}

// MakeAddressPostEndpoint returns an endpoint via the given service.
func MakeAddressPostEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(addressPostRequest)
		id, err := s.PostAddress(req.Address, req.UserID)
		return postResponse{ID: id}, err
	}
}

// MakeLoginEndpoint returns an endpoint via the given service.
func MakeDeleteEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteRequest)
		if req.AddID != "" {
			err := s.DeleteAddress(req.AddID, req.UserID)
			if err == nil {
				return statusResponse{Status: true}, err
			}
			return statusResponse{Status: false}, err
		}
		err = s.DeleteUser(req.UserID)
		if err == nil {
			return statusResponse{Status: true}, err
		}
		return statusResponse{Status: false}, err
	}
}

type GetRequest struct {
	ID   string
	Attr string
}

type loginRequest struct {
	Username string
	Password string
}

type userResponse struct {
	User dbOperations.User `json:"user"`
}

type usersResponse struct {
	Users []dbOperations.User `json:"customer"`
}

type addressPostRequest struct {
	dbOperations.Address
	UserID string `json:"userID"`
}

type addressesResponse struct {
	Addresses []dbOperations.Address `json:"address"`
}

type registerRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
}

type statusResponse struct {
	Status bool `json:"status"`
}

type postResponse struct {
	ID string `json:"id"`
}

type deleteRequest struct {
	UserID string
	AddID  string
}

type healthRequest struct {
	//
}

type EmbedStruct struct {
	Embed interface{} `json:"_embedded"`
}
