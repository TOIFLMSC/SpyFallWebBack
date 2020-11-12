// package main

// import (
// 	"flag"
// 	"log"

// 	"github.com/BurntSushi/toml"

// 	"github.com/TOIFLMSC/spyfall-web-backend/internal/app/apiserver"
// )

// var (
// 	configPath string
// )

// func init () {
// 	flag.StringVar(&configPath, "config-path", "configs/apiserver.toml", "path to config file")
// }

// func main() {
// 	flag.Parse()

// 	config := apiserver.NewConfig()

// 	_, err := toml.DecodeFile(configPath, config)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	if err := apiserver.Start(config); err != nil {
// 		log.Fatal(err)
// 	}
// router := mux.NewRouter()

// router.HandleFunc("/play/newlobby", controllers.CreateLobby).Methods("POST")
// router.HandleFunc("/play/connect/{token}/lobby", controllers.ConnectLobby).Methods("GET")

// port := os.Getenv("PORT")
// if port == "" {
// 	port = "8000"
// }

// err := http.ListenAndServe(":"+port, router)
// if err != nil {
// 	fmt.Print(err)
// } else {
// 	fmt.Println("Server is working on port ", port)
// }
//}