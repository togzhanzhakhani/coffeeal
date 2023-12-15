package models

type User struct {
    ID        uint       `gorm:"primary_key"`
    Name      string     `gorm:"not null"`
    Email     string     `gorm:"unique;not null"`
    Password  string     `gorm:"not null"`
    Orders    []Order    `gorm:"foreignkey:UserID"`
}

type Product struct {
    ID          uint         `gorm:"primary_key"`
    Name        string       `gorm:"not null"`
    Description string       `gorm:"not null"`
    Price       uint         `gorm:"not null"`
    ImageURL    string       `gorm:"not null"`
    Rating      float64         `gorm:"default:0"`
    Orders      []OrderItem  `gorm:"foreignkey:ProductID"`
}

type Cart struct {
    ID          uint         `gorm:"primary_key"`
    UserID      uint         `gorm:"not null"`
    CartItems   []CartItem   `gorm:"foreignkey:CartID"`
}

type CartItem struct {
    ID         uint        `gorm:"primary_key"`
    CartID     uint        `gorm:"not null"`
    ProductID  uint        `gorm:"not null"`
    Quantity   uint        `gorm:"not null"`
    Product 
}

type Order struct {
    ID             uint           `gorm:"primary_key"`
    UserID         uint           `gorm:"not null"`
    CartID         uint           `gorm:"not null"`
    DeliveryMethod string         `gorm:"not null"`
    PhoneNumber    string         `gorm:"not null"`
    Address        string         `gorm:"not null"`
    Cost           int            `gorm:"not null"`
    Items          []OrderItem    `gorm:"foreignkey:OrderID"`
}

type OrderItem struct {
    ID        uint `gorm:"primary_key"`
    OrderID   uint `gorm:"not null"`
    ProductID uint `gorm:"not null"`
    Quantity  uint `gorm:"not null"`
}

type RatingUpdate struct {
    Rating string `json:"rating"`
}

