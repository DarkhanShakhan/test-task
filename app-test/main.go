package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"

	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	uri             = "mongodb://root:example@localhost:27017/"
	ErrUserNotExist = errors.New("user with a given email doesn't exist")
)

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	handler := NewHandler(client)
	r := chi.NewRouter()
	r.Post("/create-user", handler.CreateUserHandler)
	r.Get("/get-user/{email}", handler.GetUserHandler)
	log.Println("Listening on port: 8989")
	if err = http.ListenAndServe(":8989", r); err != nil {
		log.Fatal(err)
	}
}

//Handler

type Handler struct {
	mongodb *mongo.Client
}

func NewHandler(mongodb *mongo.Client) *Handler {
	return &Handler{mongodb: mongodb}
}

func (h *Handler) storeUser(creds Credentials) error {
	coll := h.mongodb.Database("test").Collection("users")
	_, err := coll.InsertOne(context.TODO(), creds)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	creds := Credentials{}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !creds.validateEmail() || creds.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = h.getUserByEmail(creds.Email)
	if err == ErrUserNotExist {
		if err = creds.hashPassword(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err = h.storeUser(creds); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}

func (h *Handler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	creds, err := h.getUserByEmail(email)
	if err == ErrUserNotExist {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	payload, err := json.Marshal(creds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}

func (h *Handler) getUserByEmail(email string) (Credentials, error) {
	coll := h.mongodb.Database("test").Collection("users")
	filter := bson.D{{"email", email}}
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return Credentials{}, err
	}
	result := Credentials{}
	if cursor.Next(context.TODO()) {
		if err = cursor.Decode(&result); err != nil {
			return Credentials{}, err
		}
		return result, nil
	}
	return Credentials{}, ErrUserNotExist
}

//credentials
type Credentials struct {
	Email    string `json:"email,omitempty"`
	Salt     string `json:"salt,omitempty"`
	Password string `json:"password,omitempty"`
}

func (c *Credentials) validateEmail() bool {
	return regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`).MatchString(c.Email)
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
