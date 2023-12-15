package main

import (
	"encoding/json"
	"html/template"
	"log"

	//"log"
	"net/http"
	"strconv"
	"strings"

	"coffee.net/models"
)



func RegisterUser(w http.ResponseWriter, r *http.Request) {
	db, err := GetDB()
	if err != nil {
		log.Print(err.Error)
	}
	if r.Method == "GET" {
		http.ServeFile(w, r, "templates/register.html")
		return
	}

	if r.Method == "POST" {
		var registerUser models.User
		err := json.NewDecoder(r.Body).Decode(&registerUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("user", registerUser);

		user := models.User{
			Name:     registerUser.Name,
			Email:    registerUser.Email,
			Password: registerUser.Password,
		}

		if err := db.Create(&user).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	db, err := GetDB()
	if err != nil {
		log.Print(err.Error)
	}

	if r.Method == "GET" {
		http.ServeFile(w, r, "templates/login.html")
		return
	}

	if r.Method == "POST" {
		var loginUser models.User
		err := json.NewDecoder(r.Body).Decode(&loginUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var user models.User
		db.Where(&models.User{Email: loginUser.Email, Password: loginUser.Password}).First(&user)
		if user.ID == 0 {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func User(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/user.html")
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	db, err := GetDB()
	if err != nil {
		log.Print(err.Error)
	}

	params := strings.Split(r.URL.Path, "/")
	userID, err := strconv.Atoi(params[len(params)-1])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user models.User
	result := db.First(&user, userID)
	if result.Error != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		tmpl, err := template.ParseFiles("templates/userdetail.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")
		user.Name = name
		user.Email = email
		user.Password = password
		db.Save(&user)
		http.Redirect(w, r,"/user", http.StatusFound)
	case http.MethodDelete:
		db.Delete(&user)
		http.Redirect(w, r, "/login", http.StatusFound)
	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
