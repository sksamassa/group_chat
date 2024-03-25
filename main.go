package main

import (
	"controllers"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"migrations"
	"models"
	"net/http"
	"os"
	"templates"
	"views"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis"
	"github.com/gorilla/csrf"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Setup the database
	db, err := models.Open(models.DefaultPostgresConfig())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Run the migrations
	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// Setup the services
	userService := &models.UserService{
		DB: db,
	}
	sessionService := &models.SessionService{
		DB: db,
	}

	// Setup middlewares
	csrfMw := csrf.Protect([]byte("cc75a9e6c5b8b6a22237e6a3f902e7f2"), csrf.Secure(false))
	umw := controllers.UserMiddleware{
		SessionService: sessionService,
	}

	// Setup controllers
	usersCtrs := controllers.Users{
		UserService:    userService,
		SessionService: sessionService,
	}
	usersCtrs.Templates.SignUp = views.Must(
		views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))
	usersCtrs.Templates.SignIn = views.Must(
		views.ParseFS(templates.FS, "signin.gohtml", "tailwind.gohtml"))

	// Setup router and routes
	r := chi.NewRouter()
	r.Use(csrfMw)
	r.Use(umw.SetUser)
	// Middleware stack to enhancing the routing
	r.Use(middleware.Logger)    // Log infos about incoming request
	r.Use(middleware.RealIP)    // Access the real IP address incoming request
	r.Use(middleware.RequestID) // Generate an ID for each incoming request
	r.Use(middleware.Recoverer) // Allows to recovering from errors

	// Websocket
	redisURL := os.Getenv("REDIS_URL")
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}
	rdb = redis.NewClient(opt)
	// Serve static files from the "public" directory
	// fs := http.FileServer(http.Dir("./public"))
	// r.Handle("/*", http.StripPrefix("/", fs))
	r.HandleFunc("/websocket", handleConnections)
	go handleMessages()

	// routes setup
	r.Get("/chat_room", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "chat.gohtml", "tailwind.gohtml"))))
	r.Get("/signup", usersCtrs.SignUp)
	r.Post("/signup", usersCtrs.ProcessSignUp)
	r.Get("/signin", usersCtrs.SignIn)
	r.Post("/signin", usersCtrs.ProcessSignIn)
	r.Post("/signout", usersCtrs.ProcessSignOut)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "404 - Page Not Found")
	})

	if err := http.ListenAndServe(":4004", r); err != nil {
		panic(err)
	}
}

// websocket implementation
type ChatMessage struct {
	Username string `json:"username"`
	Text     string `json:"text"`
}

var (
	rdb *redis.Client
)

var clients = make(map[*websocket.Conn]bool)
var broadcaster = make(chan ChatMessage)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// ensure connection close when function returns
	defer ws.Close()
	clients[ws] = true

	// if it's zero, no messages were ever sent/saved
	if rdb.Exists("chat_messages").Val() != 0 {
		sendPreviousMessages(ws)
	}

	for {
		var msg ChatMessage
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			delete(clients, ws)
			break
		}
		// send new message to the channel
		broadcaster <- msg
	}
}

func sendPreviousMessages(ws *websocket.Conn) {
	chatMessages, err := rdb.LRange("chat_messages", 0, -1).Result()
	if err != nil {
		panic(err)
	}

	// send previous messages
	for _, chatMessage := range chatMessages {
		var msg ChatMessage
		json.Unmarshal([]byte(chatMessage), &msg)
		messageClient(ws, msg)
	}
}

// If a message is sent while a client is closing, ignore the error
func unsafeError(err error) bool {
	return !websocket.IsCloseError(err, websocket.CloseGoingAway) && err != io.EOF
}

func handleMessages() {
	for {
		// grab any next message from channel
		msg := <-broadcaster

		storeInRedis(msg)
		messageClients(msg)
	}
}

func storeInRedis(msg ChatMessage) {
	json, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	if err := rdb.RPush("chat_messages", json).Err(); err != nil {
		panic(err)
	}
}

func messageClients(msg ChatMessage) {
	// send to every client currently connected
	for client := range clients {
		messageClient(client, msg)
	}
}

func messageClient(client *websocket.Conn, msg ChatMessage) {
	err := client.WriteJSON(msg)
	if err != nil && unsafeError(err) {
		log.Printf("error: %v", err)
		client.Close()
		delete(clients, client)
	}
}
