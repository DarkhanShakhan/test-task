package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()
	r.Post("/create-user", CreateUserHandler)
	r.Get("/get-user/{email}", GetUserHandler)

	c := Credentials{Email: "darkhan", Password: "hello"}
	c.hashPassword()
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	creds := Credentials{}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !creds.validateEmail() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//TODO: check for existing user email
	if err = creds.hashPassword(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//TODO:store user credentials
	w.WriteHeader(http.StatusCreated)
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: get user from repo
}

//credentials
type Credentials struct {
	Email    string `json:"email,omitempty"`
	Salt     string `json:"salt,omitempty"`
	Password string `json:"password,omitempty"`
}

func (c *Credentials) validateEmail() bool {
	return true
}

func (c *Credentials) hashPassword() error {
	if err := c.getSalt(); err != nil {
		return err
	}
	h := md5.New()
	h.Write([]byte(c.Password))
	c.Password = hex.EncodeToString(h.Sum([]byte(c.Salt)))
	return nil
}

func (c *Credentials) getSalt() error {
	resp, err := http.Get("http://localhost:8787/generate-salt")
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		temp := Credentials{}
		err = json.NewDecoder(resp.Body).Decode(&temp)
		if err != nil {
			return err
		}
		c.Salt = temp.Salt
		return nil
	}
	return errors.New("internal server error")
}
