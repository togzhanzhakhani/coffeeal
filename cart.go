package main

import(
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"gorm.io/gorm"

	"coffee.net/models"
)

type QuantityJSON struct {
	Quantity string `json:"quantity"`
}

func CartHandler(w http.ResponseWriter, r *http.Request) {
	db, err := GetDB()
	if err!=nil{
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
        return
	}
	params := strings.Split(r.URL.Path, "/")
	userid, err := strconv.Atoi(params[2])
	if err != nil{
		log.Print(err.Error())
		w.WriteHeader(http.StatusBadRequest)
        return
	}
	
	switch r.Method{
	case http.MethodPost:
		productid, err := strconv.Atoi(params[3])
		if err != nil{
			log.Println(err.Error())
			w.WriteHeader(http.StatusBadRequest)
            return
		}
		err = addProductToCart(userid, productid, r, db)
		if err != nil{
			log.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
            return
		} 

	case http.MethodGet:
		cartItems, err := getCartItemsByUserID(uint(userid), db)
		if err != nil {
			log.Println(err.Error())
		}
		for i, item := range cartItems {
			var product models.Product
			db.Where("id = ?", item.ProductID).First(&product)
			cartItems[i].Product = product
		}
	
		tmpl, err := template.ParseFiles("templates/cart.html")
		if err != nil {
			log.Println(err.Error())
			return
		}
		err = tmpl.Execute(w, cartItems)
		if err != nil {
			log.Println(err.Error())
			return
		}

	case http.MethodPut:
		productid, err := strconv.Atoi(params[3])
		if err != nil{
			log.Println(err.Error())
			w.WriteHeader(http.StatusBadRequest)
            return
		}
		err = updateQuantity(userid, productid, r, db)
		if err != nil{
			log.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
            return
		} 
	case http.MethodDelete:
		productid, err := strconv.Atoi(params[3])
		if err != nil{
			log.Println(err.Error())
			w.WriteHeader(http.StatusBadRequest)
            return
		}
		err = deleteCartItem(userid, productid, r, db)
		if err != nil{
			log.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
            return
		} 
	}
	
}

func getCartItemsByUserID(userId uint, db *gorm.DB) ([]models.CartItem, error) {
    var cart models.Cart
    err := db.Where("user_id = ?", userId).Preload("CartItems").First(&cart).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            log.Println("Корзина для пользователя не найдена", userId)
        } else {
            log.Println("Ошибка при запросе к базе данных", err)
        }
        return nil, err
    }
    return cart.CartItems, nil
}


func addProductToCart(userId int, productId int, r *http.Request, db *gorm.DB) error {	
	var quantityJSON QuantityJSON
	json.NewDecoder(r.Body).Decode(&quantityJSON)
	
	quantity, err := strconv.ParseUint(quantityJSON.Quantity, 10, 32)
	if err != nil {
		log.Println(err.Error())
	}
	
	cartItem := models.CartItem{
		Quantity: uint(quantity),
	}


    var cart models.Cart
    err = db.Where("user_id = ?", userId).Preload("CartItems").First(&cart).Error
    if err != nil && err != gorm.ErrRecordNotFound {
        cart.ID = 0
    }

    if cart.ID == 0 {
        newCartItem := &models.CartItem{
            ProductID: uint(productId),
            Quantity:  cartItem.Quantity,
        }
        newCart := &models.Cart{
            UserID:    uint(userId),
            CartItems: []models.CartItem{*newCartItem},
        }
        err = db.Create(newCart).Error
        if err != nil {
            log.Println(err.Error())
            return err
        }
    } else {
        var existingCartItem models.CartItem
        db.Where("product_id = ? AND cart_id = ?", productId, cart.ID).FirstOrCreate(&existingCartItem)
        if existingCartItem.ID == 0 {
            newCartItem := &models.CartItem{
                ProductID: uint(productId),
                Quantity:  cartItem.Quantity,
                CartID:    cart.ID,
            }
            err = db.Create(newCartItem).Error
            if err != nil {
                log.Println(err.Error())
                return err
            }
        } else {
            existingCartItem.Quantity += cartItem.Quantity
            err = db.Save(&existingCartItem).Error
            if err != nil {
                log.Println(err.Error())
                return err
            }
        }
    }

    return nil
}

func updateQuantity(userId int, productId int, r *http.Request, db *gorm.DB) error {
	type Quantityjs struct {
		Quantity int `json:"quantity"`
	}	
	var quantityJSON Quantityjs
	json.NewDecoder(r.Body).Decode(&quantityJSON)

    cartItem := models.CartItem{}
    if err := db.Where("product_id = ?", productId).First(&cartItem).Error; err != nil {
        log.Println(err.Error())
        return err
    }
    cartItem.Quantity = uint(quantityJSON.Quantity)

    if err := db.Save(&cartItem).Error; err != nil {
        log.Println(err.Error())
        return err
    }
	return nil
}

func deleteCartItem(userId int, productId int, r *http.Request, db *gorm.DB) error {
    var cart models.Cart
    err := db.Where("user_id = ?", userId).Preload("CartItems").First(&cart).Error
    if err != nil {
        log.Println(err.Error())
        return err
    }

    var cartItem models.CartItem
    for i, item := range cart.CartItems {
        if item.ProductID == uint(productId) {
			cartItem = item
            err = db.Delete(&item).Error
			if err != nil {
				log.Println(err.Error())
				return err
			}
			cart.CartItems = append(cart.CartItems[:i], cart.CartItems[i+1:]...)
        }
    }

    if cartItem.ID == 0 {
        log.Println(err.Error())
        return err
    }

    err = db.Save(&cart).Error
    if err != nil {
        log.Println(err.Error())
        return err
    }
	return nil
}
