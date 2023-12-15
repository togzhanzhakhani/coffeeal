package main
import (
	"encoding/json"
	//"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	//"math"
	"gorm.io/gorm"

	"coffee.net/models"
)

func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := GetDB()
	if err!=nil {
		log.Print(err.Error())
	}

	switch r.Method{
	case http.MethodGet:
		params := strings.Split(r.URL.Path, "/")
		userid, err := strconv.Atoi(params[2])
		if err != nil{
			log.Print("userid", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		orders := []models.Order{}

		if err := db.Where("user_id = ?", userid).Find(&orders).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("templates/orders.html")
		if err != nil {
			log.Println(err.Error())
			return
		}
		err = tmpl.Execute(w, orders)
		if err != nil {
			log.Println(err.Error())
			return
		}

	}
}


type orderJSON struct{
    UserID         string           `json:"UserID"`
    DeliveryMethod string           `json:"DeliveryMethod"`
    PhoneNumber    string           `json:"PhoneNumber"`
    Address        string           `json:"Address"`
    Cost           string           `json:"Cost"`
}

func PostOrder(w http.ResponseWriter, r *http.Request) {
	db, err := GetDB()
	if err!=nil {
		log.Print(err.Error())
	}
	switch r.Method{
	case http.MethodPost:
		err = postOrder(r, db)
		if err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	
}

func postOrder(r *http.Request, db *gorm.DB) error {
	var orderJSON orderJSON
	json.NewDecoder(r.Body).Decode(&orderJSON)
	
	userid, err := strconv.ParseUint(orderJSON.UserID, 10, 32)
	if err != nil {
		log.Println("user", err.Error())
	}

	cost, err := strconv.Atoi(orderJSON.Cost)
	if err != nil {
		log.Println("user", err.Error())
	}

	var cart models.Cart
    err = db.Where("user_id = ?", userid).Preload("CartItems").First(&cart).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            log.Println("Корзина для пользователя не найдена", userid)
        } else {
            log.Println("Ошибка при запросе к базе данных", err)
        }
        return err
    }
	
	newOrder := models.Order{
		UserID: uint(userid),
		CartID: uint(cart.ID),
		DeliveryMethod: orderJSON.DeliveryMethod,
		PhoneNumber: orderJSON.PhoneNumber,
		Address: orderJSON.Address,
		Cost: cost,
	}
    
    defer r.Body.Close()

    db.Create(&newOrder)

    for _, cartItem := range cart.CartItems {
        orderItem := models.OrderItem{
            OrderID:   newOrder.ID,
            ProductID: cartItem.ProductID,
            Quantity:  cartItem.Quantity,
        }
        db.Create(&orderItem)
		newOrder.Items = append(newOrder.Items, orderItem)
    }
	db.Save(&newOrder)
	if err := db.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
        db.Rollback()
        return err
    }
	db.Delete(&cart)
	return nil
}
