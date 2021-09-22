package main
// 1
import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)
// 2
var rnd *renderer.Render
var db *mgo.Database
// 3
const (
	hostName string = "localhost:27017"
	dbName string = "demo_todo"
	collectionName string = "todo"
	port string = ":9000"
)
// 4
type(
	todoModel struct {
		ID bson.ObjectId `bson:"id, omitempty"`
		Title string `bson:"title"`
		Completed bool `bson:"completed"`
		CreatedAt time.Time `bson:"created_at"`
	}

	todo struct {
		ID string `bson:"id"`
		Title string `bson:"title"`
		Completed bool `bson:"completed"`
		CreatedAt time.Time `bson:"created_at"`
	}
)
// 5
func init()  {
	rnd = renderer.New()
	sess, err := mgo.Dial(hostName)
	checkErr(err)
	sess.SetMode(mgo.Monotonic, true)
	db = sess.DB(dbName)
}
// 6
func checkErr(err error)  {
	if err != nil{
		log.Fatal(err)
	}
}

func main() {
// 11
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)
// 7
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homeHandler)
	r.Mount("/todo", todoHandlers())
// 8
	srv := &http.Server{
		Addr: port,
		Handler: r,
		ReadTimeout: 60*time.Second,
		WriteTimeout: 60*time.Second,
		IdleTimeout: 60*time.Second,
	}
// 9
	go func() {
		log.Println("Listening in port ", port)
		if err := srv.ListenAndServe(); err != nil{
			log.Println("Listen: %s\n", err)
		}
	}()
// 12
	<-stopChan
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel()
		log.Println("Server gracefully stopped!")
}
// 10
func todoHandlers() http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/", fetchTodos)
		r.Post("/", createTodo)
		r.Put("/{id}", updateTodo)
		r.Delete("/{id}", deleteTodo)
	})
	return rg
}
// 13
func homeHandler(w http.ResponseWriter, r *http.Request)  {
	err := rnd.Template(w, http.StatusOK, []string{"static/home.tpl"}, nil)
	checkErr(err)
}
// 14
func fetchTodos(w http.ResponseWriter, r *http.Request)  {
	todos := []todoModel{}

	if err := db.C(collectionName).Find(bson.M{}).All(&todos); err != nil{
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to fecth todo",
			"error": err,
		})
		return
	}

	todoList := []todo{}

	for _, t := range todos{
		todoList = append(todoList, todo{
			ID: t.ID.Hex(),
			Title: t.Title,
			Completed: t.Completed,
			CreatedAt: t.CreatedAt,
		})
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"data": todoList,
	})
}
// 15
func createTodo(w http.ResponseWriter, r *http.Request)  {
	var t todo

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil{
		rnd.JSON(w, http.StatusProcessing, err)
		return
	}

	if t.Title == ""{
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title field is requried",
		})
		return
	}

	tm := todoModel{
		ID: bson.NewObjectId(),
		Title: t.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}

	if err := db.C(dbName).Insert(&tm); err != nil{
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to save todo",
			"error": err,
		})
		return
	}

	rnd.JSON(w, http.StatusCreated, renderer.M{
		"message": "Todo created successfully",
		"todo_id": tm.ID.Hex(),
	})
}
// 16
func updateTodo(w http.ResponseWriter, r *http.Request)  {
	id := strings.TrimSpace(chi.URLParam(r, "id"))

	if !bson.IsObjectIdHex(id){
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title field is requried",
		})
		return
	}

	var t todo

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil{
		rnd.JSON(w, http.StatusProcessing, err)
		return
	}

	if t.Title == ""{
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title field is requried",
		})
		return
	}

	if err := db.C(collectionName).Update(
			bson.M{
				"_id": bson.ObjectIdHex(id),
			},
			bson.M{
				"title": t.Title,
				"completed": t.Completed,
			},
			); err != nil{
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to update todo",
			"error": err,
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo update successfully",
	})
}
// 17
func deleteTodo(w http.ResponseWriter, r *http.Request)  {
	id := strings.TrimSpace(chi.URLParam(r, "id"))

	if !bson.IsObjectIdHex(id){
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo deleted successfully",
	})
}
