package main

import (
	"log"
	"net/http"

	"coffee.net/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/lib/pq"
)


func main() {
	db, err := GetDB()
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&models.User{}, &models.Product{}, &models.Order{}, &models.OrderItem{}, &models.Cart{}, &models.CartItem{})
	if err != nil {
		log.Fatal(err)
	}

    if err != nil {
        panic(err)
    }
	//app := handlers.NewApp(dbs)

	http.HandleFunc("/register", RegisterUser)
	http.HandleFunc("/login", LoginUser)
	http.HandleFunc("/user", User)
	http.HandleFunc("/user/", UserHandler)
	http.HandleFunc("/products", GetProductsHandler)
	http.HandleFunc("/products/", GetProductHandler)
	http.HandleFunc("/products/manage/", ManageProductsHandler)
	http.HandleFunc("/cart/", CartHandler)
	http.HandleFunc("/orders", PostOrder)
	http.HandleFunc("/orders/", GetOrdersHandler)

	log.Println("Starting server on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}


func GetDB() (*gorm.DB, error) {
	dsn := "host=database user=postgres password=postgres dbname=coffee-shop port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    return db, err
}