package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	database "github.com/anuj-jha-work/rss_Agregator/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	//apne db ka pass dia gya hai env me kyunki vo sirf system me hi rahna chahiye isliye env me rakha gya hai
	godotenv.Load(".env")

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT,is not Working Please Try Again")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URl is not Working Please Try Again")
	}
//ek connection pool banta hai jisse har new req ke liye alag connection na bna pade aur vo re use ho sake
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("DB_URl is not Working Please Try Again")
	}
// isi pool ke connection ko use karke queries run karenge
	apiCfg := apiConfig{
		DB: database.New(conn),
	}
	go startScraping(apiCfg.DB, 10, time.Minute)
	//jab koi brwoser ya thunderclients e req karta hua to server ko kes epata lega ki konsa fn chalana hai iske liye hum chi router ka use karenge
	router := chi.NewRouter()
	//browser a security rul hai lag alg ports  se req ayga to block kardega cors laga ke bolte hai agr alg alg ya kahi s ehia ey to aloow karna
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	//humne versioning ke liye v1Router banaya hai taki future me v2 ya v3 aayega to easily manage ho sake aur old users  ka work pahtega nhi
	v1Router := chi.NewRouter()

	v1Router.Post("/users", apiCfg.handlerCreateUser)
	v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))

	v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
	v1Router.Get("/feeds", apiCfg.handlerGetFeeds)

	v1Router.Get("/posts", apiCfg.middlewareAuth(apiCfg.handlerGetPostsForUser))

	v1Router.Post("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
	v1Router.Get("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollows))
	v1Router.Delete("/feed_follows/{feedFollowID}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollow))

	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/err", handlerError)

	router.Mount("/v1", v1Router)
	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	log.Printf("Serving on port: %s\n", portString)
	log.Fatal(srv.ListenAndServe())
}
