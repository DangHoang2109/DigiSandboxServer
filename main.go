package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Thread struct {
	ID          uuid.UUID `db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
}

type Post struct {
	ID       uuid.UUID `db:"id"`
	ThreadID uuid.UUID `db:"thread_id"`
	Title    string    `db:"title"`
	Content  string    `db:"content"`
	Votes    int       `db:"votes"`
}

type Comment struct {
	ID      uuid.UUID `db:"id"`
	PostID  uuid.UUID `db:"post_id"`
	Content string    `db:"content"`
	Votes   int       `db:"votes"`
}

type Person struct {
	ID        primitive.ObjectID `json:"_id, omitempty", bson:"_id, omitempty"`
	FirstName string             `json:"firstname, omitempty", bson:"firstname, omitempty"`
	LastName  string             `json:"lastname, omitempty", bson:"lastname, omitempty"`
}

type ThreadStore interface {
	//Get một thread bằng id
	Thread(id uuid.UUID) (Thread, error)
	//Get tất cả threads
	Threads() ([]Thread, error)
	//Create một Thread t
	CreateThread(t *Thread) error
	//Update một thread t
	UpdateThread(t *Thread) error
	//Delte một thread có id
	DeleteThread(id uuid.UUID) error
}

type PostStore interface {
	//Get một Post bằng id
	Post(id uuid.UUID) (Post, error)
	//Get tất cả Posts trong thread có id
	Posts(threadID uuid.UUID) ([]Post, error)
	//Create một Post t
	CreatePost(t *Post) error
	//Update một Post t
	UpdatePost(t *Post) error
	//Delte một Post có id
	DeletePost(id uuid.UUID) error
}

type CommentStore interface {
	//Get một Comment bằng id
	Comment(id uuid.UUID) (Comment, error)
	//Get tất cả Comments trong Thread id
	Comments(threadID uuid.UUID) ([]Comment, error)
	//Create một Comment t
	CreateComment(t *Comment) error
	//Update một Comment t
	UpdateComment(t *Comment) error
	//Delte một Comment có id
	DeleteComment(id uuid.UUID) error
}

//Create a instance for connection
func main() {
	///Connect mongo
	//client := ConnectMongo()

	uri := "mongodb+srv://admin:Thuytrang2109@cluster0.yzwud.gcp.mongodb.net/LoginServer?retryWrites=true&w=majority"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			fmt.Println("Disconnect.")
			panic(err)
		}
	}()
	// Ping the primary
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		fmt.Println("Eror.")
		panic(err)
	}
	fmt.Println("Successfully connected and pinged.")

	//Handle router
	InitRouter(client)
}

func InitRouter(client *mongo.Client) {

	router := mux.NewRouter()

	router.HandleFunc("/hello", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Heeloo"))
	})

	router.HandleFunc("/create-person", func(rw http.ResponseWriter, r *http.Request) {
		CreatePersonEndpoint(rw, r, client)
	}).Methods("POST")
	router.HandleFunc("/persons", func(rw http.ResponseWriter, r *http.Request) {
		GetPersonsEndPoint(rw, r, client)
	}).Methods("GET")
	router.HandleFunc("/person/{name}", func(rw http.ResponseWriter, r *http.Request) {
		GetPersonEndPoint(rw, r, client)
	}).Methods("GET")

	http.ListenAndServe(":8080", router)
}

func CreatePersonEndpoint(rw http.ResponseWriter, r *http.Request, client *mongo.Client) {
	rw.Header().Add("content-type", "application/json")
	var person Person
	json.NewDecoder(r.Body).Decode(&person)

	collection := client.Database("LoginServer").Collection("people")
	if collection == nil {
		fmt.Println("Eror.")
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, _ := collection.InsertOne(ctx, person)
	json.NewEncoder(rw).Encode(result)
}
func GetPersonEndPoint(rw http.ResponseWriter, r *http.Request, client *mongo.Client) {
	rw.Header().Add("content-type", "application/json")

	params := mux.Vars(r)
	id := params["name"] //primitive.ObjectIDFromHex(params["id"])

	var person Person

	collection := client.Database("LoginServer").Collection("people")
	if collection == nil {
		fmt.Println("Eror.")
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err := collection.FindOne(ctx, Person{FirstName: id})
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(`{"msg:  "` + err.Err().Error() + `"}"`))
		return
	}

	json.NewEncoder(rw).Encode(person)
}
func GetPersonsEndPoint(rw http.ResponseWriter, r *http.Request, client *mongo.Client) {
	rw.Header().Add("content-type", "application/json")
	var persons []Person

	collection := client.Database("LoginServer").Collection("people")
	if collection == nil {
		fmt.Println("Eror.")
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	curser, err := collection.Find(ctx, bson.M{})
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(`{"msg:  "` + err.Error() + `"}"`))
		return
	}
	defer curser.Close(ctx)

	for curser.Next(ctx) {
		var person Person
		curser.Decode(&person)
		persons = append(persons, person)
	}

	json.NewEncoder(rw).Encode(persons)
}
func ConnectMongo() *mongo.Client {
	// Replace the uri string with your MongoDB deployment's connection string.
	uri := "mongodb+srv://admin:Thuytrang2109@cluster0.yzwud.gcp.mongodb.net/LoginServer?retryWrites=true&w=majority"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			fmt.Println("Disconnect.")
			panic(err)
		}
	}()
	// Ping the primary
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		fmt.Println("Eror.")
		panic(err)
	}
	fmt.Println("Successfully connected and pinged.")

	return client
}
