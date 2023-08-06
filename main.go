package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	db     *gorm.DB
	secret = "your-secret-key"
)

type Product struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Category    string `json:"category"`
	Quantity    int    `json:"quantity"`
}

type Customer struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"password"`
}

type ShoppingCart struct {
	gorm.Model
	CustomerID uint
	ProductID  uint
	Quantity   int
}

type Order struct {
	gorm.Model
	CustomerID  uint
	ProductID   uint
	Quantity    int
	TotalAmount string
}

type JWTClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func main() {
	initDB()
	defer db.Close()

	app := fiber.New()

	app.Get("/api/products", getProductList)
	app.Get("/api/products/:id", getProduct)
	app.Get("/api/shopping-cart", viewShoppingCart)
	app.Post("/api/shopping-cart", addToShoppingCart)
	app.Delete("/api/shopping-cart/:id", deleteFromShoppingCart)
	app.Post("/api/checkout", checkout)
	app.Post("/api/register", registerCustomer)
	app.Post("/api/login", login)

	fmt.Println("Server is running on :8080")
	log.Fatal(app.Listen(":8080"))
}

func initDB() {
	dsn := "user:password@tcp(localhost:3306)/online_store?charset=utf8&parseTime=True"
	conn, err := gorm.Open("mysql", dsn)
	if err != nil {
		panic("failed to connect database")
	}

	db = conn
	db.AutoMigrate(&Product{}, &Customer{}, &ShoppingCart{}, &Order{})
}

func getProductList(c *fiber.Ctx) error {
	var products []Product
	db.Find(&products)
	return c.JSON(products)
}

func getProduct(c *fiber.Ctx) error {
	productID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	var product Product
	if err := db.First(&product, productID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	return c.JSON(product)
}

func viewShoppingCart(c *fiber.Ctx) error {
	user, err := getUserFromToken(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var cartItems []ShoppingCart
	if err := db.Where("customer_id = ?", user.ID).Find(&cartItems).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch shopping cart items",
		})
	}

	return c.JSON(cartItems)
}

func addToShoppingCart(c *fiber.Ctx) error {
	user, err := getUserFromToken(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var cartItem ShoppingCart
	if err := c.BodyParser(&cartItem); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	cartItem.CustomerID = user.ID
	if err := db.Create(&cartItem).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add item to shopping cart",
		})
	}

	return c.Status(http.StatusCreated).JSON(cartItem)
}

func deleteFromShoppingCart(c *fiber.Ctx) error {
	user, err := getUserFromToken(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	cartItemID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid cart item ID",
		})
	}

	var cartItem ShoppingCart
	if err := db.First(&cartItem, cartItemID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Cart item not found",
		})
	}

	if cartItem.CustomerID != user.ID {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to delete this cart item",
		})
	}

	if err := db.Delete(&cartItem).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete item from shopping cart",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Cart item deleted",
	})
}

func checkout(c *fiber.Ctx) error {
	user, err := getUserFromToken(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var cartItems []ShoppingCart
	if err := db.Where("customer_id = ?", user.ID).Find(&cartItems).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch shopping cart items",
		})
	}

	var totalPrice float64
	for _, item := range cartItems {
		var product Product
		if err := db.First(&product, item.ProductID).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch product details",
			})
		}
		price, _ := strconv.ParseFloat(product.Price, 64)
		totalPrice += price * float64(item.Quantity)
	}

	var userBalance float64
	if userBalance < totalPrice {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Insufficient balance",
		})
	}

	for _, item := range cartItems {
		var product Product
		if err := db.First(&product, item.ProductID).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch product details",
			})
		}

		if product.Quantity < item.Quantity {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Not enough stock for some products",
			})
		}

		product.Quantity -= item.Quantity
		db.Save(&product)

		order := Order{
			CustomerID:  user.ID,
			ProductID:   item.ProductID,
			Quantity:    item.Quantity,
			TotalAmount: product.Price,
		}
		if err := db.Create(&order).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create order record",
			})
		}
	}

	userBalance -= totalPrice

	if err := db.Where("customer_id = ?", user.ID).Delete(&ShoppingCart{}).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to clear shopping cart",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Checkout completed",
	})
}

func registerCustomer(c *fiber.Ctx) error {
	var customer Customer
	if err := c.BodyParser(&customer); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	var existingCustomer Customer
	if err := db.Where("username = ?", customer.Username).First(&existingCustomer).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "Username already exists",
		})
	}

	if err := db.Create(&customer).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to register customer",
		})
	}

	return c.Status(http.StatusCreated).JSON(customer)
}

func login(c *fiber.Ctx) error {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&credentials); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	var customer Customer
	if err := db.Where("username = ? AND password = ?", credentials.Username, credentials.Password).First(&customer).Error; err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid username or password",
		})
	}

	token, err := createToken(credentials.Username)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create JWT token",
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}

func createToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &JWTClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func getUserFromToken(c *fiber.Ctx) (*Customer, error) {
	tokenString := c.Get("Authorization")
	if tokenString == "" {
		return nil, fmt.Errorf("missing authorization token")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	var user Customer
	if err := db.Where("username = ?", claims.Username).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
