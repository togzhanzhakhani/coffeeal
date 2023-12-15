package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"math"

	"coffee.net/models"

)

func GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	minPrice, err := strconv.Atoi(r.URL.Query().Get("priceMin"))
    if err != nil {
        minPrice = 0
    }
    maxPrice, err := strconv.Atoi(r.URL.Query().Get("priceMax"))
    if err != nil {
        maxPrice = math.MaxInt32
    }

	products, err := GetProducts(minPrice, maxPrice)
	if err != nil{
		log.Print(err.Error())
		return
	}

	data := struct {
		PriceMin   int
		PriceMax   int
		Products []models.Product
	}{
		PriceMin:   minPrice,
		PriceMax: maxPrice,
		Products: products,
	}

	tmpl, err := template.ParseFiles("templates/products.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}


func GetProductHandler(w http.ResponseWriter, r *http.Request) {
    db,err := GetDB()
	if err != nil{
		log.Print(err.Error())
	}	
	params := strings.Split(r.URL.Path, "/")
	id := params[len(params)-1]

	switch r.Method {
	case http.MethodGet:
		var product models.Product
		if err := db.First(&product, id).Error; err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		tmpl, err := template.ParseFiles("templates/productsdetail.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, product)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		ratingUpdate := &models.RatingUpdate{}
		err = json.NewDecoder(r.Body).Decode(ratingUpdate)
		if err != nil {
			log.Println("error", err.Error())
			return
		}

		rating, err := strconv.ParseFloat(ratingUpdate.Rating, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var product models.Product
		if err := db.Where("id = ?", id).First(&product).Error; err != nil {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}

		product.Rating += rating
		if product.Rating != rating{
			product.Rating = product.Rating / 2	
		}
		

		if err := db.Save(&product).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
	}
}

func GetProducts(minPrice, maxPrice int)([]models.Product, error){
	db,err := GetDB()
	if err != nil{
		log.Print(err.Error)
	}	

	var products []models.Product
	if err := db.Where("price >= ? AND price <= ?", minPrice, maxPrice).Find(&products).Error; err != nil {
		log.Print(err.Error)
		return nil, err
	}
	return products, err
}

func IsAdmin(r *http.Request) (bool, error) {
	params := strings.Split(r.URL.Path, "/")
    currentUserid, err := strconv.Atoi(params[3])
	if err != nil{
		log.Print(err.Error())
		return false, err
	}
    admins := []int{3}
    
    for _, adminid := range admins {
        if currentUserid == adminid {
            return true, nil
        }
    }
    return false, nil
}


func ManageProductsHandler(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path, "/")
	userID := params[3]
	if len(params)>4{
		ManageProductByIDHandler(w, r)
		return
	}
	db, err := GetDB()
	if err!=nil{
		log.Print(err.Error)
	}
    isAdmin, err := IsAdmin(r)
    if err != nil {
        http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
        return
    }
    if !isAdmin {
        http.Error(w, "Доступ запрещен", http.StatusForbidden)
        return
    }
	switch r.Method {
	case http.MethodGet:
		products, err := GetProducts(0, math.MaxInt32)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := struct {
			UserID   string
			Products []models.Product
		}{
			UserID:   userID,
			Products: products,
		}

		tmpl, err := template.ParseFiles("templates/productsmanage.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		name := r.FormValue("name")
		description := r.FormValue("description")
		price, err := strconv.ParseUint(r.FormValue("price"), 10, 64)
		if err != nil {
			log.Print(err.Error())
		}
		imageURL := r.FormValue("image_url")

		product := models.Product{
			Name:        name,
			Description: description,
			Price:       uint(price),
			ImageURL:    imageURL,
		}

		if err := db.Create(&product).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(product)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func ManageProductByIDHandler(w http.ResponseWriter, r *http.Request){
	db, err := GetDB()
	if err!=nil{
		log.Print(err.Error)
	}

    isAdmin, err := IsAdmin(r)
    if err != nil {
        http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
        return
    }

    if !isAdmin {
        http.Error(w, "Доступ запрещен", http.StatusForbidden)
        return
    }

	params := strings.Split(r.URL.Path, "/")
    productid, err := strconv.Atoi(params[4])
	if err != nil{
		log.Print(err.Error())
		return
	}

	var product models.Product
	if err := db.First(&product, productid).Error; err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	currentUserID := params[len(params)-2]
	if r.Method == http.MethodPost {
        name := r.FormValue("name")
        description := r.FormValue("description")
        priceStr := r.FormValue("price")
        price, err := strconv.ParseUint(priceStr, 10, 64)
        if err != nil {
            http.Error(w, "Invalid price", http.StatusBadRequest)
            return
        }
        imageURL := r.FormValue("image_url")

        product.Name = name
        product.Description = description
        product.Price = uint(price)
        product.ImageURL = imageURL
		db.Save(&product)

        http.Redirect(w, r, fmt.Sprintf("/products/manage/%s", currentUserID), http.StatusSeeOther)
        return
    }

    if r.Method == http.MethodDelete {
        db.Delete(&product)
        http.Redirect(w, r, fmt.Sprintf("/products/manage/%s", currentUserID), http.StatusSeeOther)
        return
    }

    tmpl, err := template.ParseFiles("templates/productsdetailmanage.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    err = tmpl.Execute(w, product)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

}

