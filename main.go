package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"golang_api/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// Struct Repository
type Repository struct {
	DB      *gorm.DB
	CartMap map[uint]int
}

// Struct Message
type Message struct {
	Message string `json:"message"`
}

// Struct Register & Log_In
type (
	Account struct {
		Fullname         string `json:"fullname"`
		Email            string `json:"email"`
		Username         string `json:"username"`
		Password         string `json:"password"`
		Confirm_Password string `json:"confirm_password"`
	}

	LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
)
type CartItem struct {
	gorm.Model
	ProductID uint
	Quantity  int
}

// Struct UpdateAccountRequest
type UpdateAccountRequest struct {
	Fullname string `json:"fullname"`
	Age      int    `json:"age"`
	Address  string `json:"address"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// Struct UpdateUserRequest (by Admin)
type UpdateUserRequest struct {
	Username string `json:"username"`
	Fullname string `json:"fullname"`
	Age      int    `json:"age"`
	Address  string `json:"address"`
	Email    string `json:"email"`
}

// Struct Change password
type UpdatePasswordRequest struct {
	Username        string `json:"username"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// Struct Loan application
type LoanApplication struct {
	FullName            string `json:"full_name"`
	Email               string `json:"email"`
	PhoneNumber         string `json:"phone_number"`
	LoanDetails         string `json:"loan_details"`
	LoanAmountRequested string `json:"loan_amount_requested"`
	PurposeOfLoan       string `json:"purpose_of_loan"`
	EmploymentDetails   string `json:"employment_details"`
	RepaymentPlan       string `json:"repayment_plan"`
}

// Struct Product
type Product struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	ImageData   []byte  `json:"image_data"`
}

type CartItems struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	Product  uint `json:"product"`
	Quantity int  `json:"quantity"`
}

// Struct GetUserDataResponse
type GetUserDataResponse struct {
	Fullname string `json:"fullname"`
	Age      int    `json:"age"`
	Address  string `json:"address"`
	Email    string `json:"email"`
}

// Struct Rider Account
type (
	RiderAccount struct {
		Fullname         string `json:"fullname"`
		Age              int    `json:"age"`
		Address          string `json:"address"`
		Email            string `json:"email"`
		Username         string `json:"username"`
		DriverLicenseNum string `json:"driver_license_num"`
		Password         string `json:"password"`
		Confirm_Password string `json:"confirm_password"`
	}
	UpdateRiderAccountRequest struct {
		Username         string `json:"username"`
		Fullname         string `json:"fullname"`
		Age              int    `json:"age"`
		Address          string `json:"address"`
		Email            string `json:"email"`
		DriverLicenseNum string `json:"driver_license_num"`
	}
)

// Struct Order
type Order struct {
	Fullname   string `json:"fullname"`
	Mobile     string `json:"mobile"`
	Address    string `json:"address"`
	ItemTitle  string `json:"itemTitle"`
	Quantity   int    `json:"quantity"`
	PurchaseID uint   `json:"-"`
}

// HASH
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// Create Account
func (r *Repository) CreateAccount(context *fiber.Ctx) error {
	account := Account{}
	err := context.BodyParser(&account)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "invalid request"})
		return err
	}

	if account.Password != account.Confirm_Password {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "passwords do not match"})
		return nil
	}

	//if the username or email already exists
	var existingAccount Account
	err = r.DB.Table("account").Where("username = ? OR email = ?", account.Username, account.Email).First(&existingAccount).Error
	if err == nil {
		context.Status(http.StatusConflict).JSON(
			&fiber.Map{"message": "username or email already exists"})
		return nil
	}

	// Hash the password
	hashedPassword, err := hashPassword(account.Password)
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "error hashing password"})
		return err
	}

	// Create the new account
	newAccount := Account{
		Fullname: account.Fullname,
		Email:    account.Email,
		Username: account.Username,
		Password: hashedPassword,
	}

	err = r.DB.Table("account").Create(&newAccount).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not create account"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Successfully Registered!!!"})
	return nil
}

// Create Loan Application
func (r *Repository) CreateLoanApplication(context *fiber.Ctx) error {
	loanapplication := LoanApplication{}
	err := context.BodyParser(&loanapplication)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "invalid request"})
		return err
	}

	// Loan Application
	newLoanApplication := LoanApplication{

		FullName:            loanapplication.FullName,
		Email:               loanapplication.Email,
		PhoneNumber:         loanapplication.PhoneNumber,
		LoanDetails:         loanapplication.LoanDetails,
		LoanAmountRequested: loanapplication.LoanAmountRequested,
		PurposeOfLoan:       loanapplication.PurposeOfLoan,
		EmploymentDetails:   loanapplication.EmploymentDetails,
		RepaymentPlan:       loanapplication.RepaymentPlan,
	}

	err = r.DB.Table("loanapplication").Create(&newLoanApplication).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not Create Loan Application"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "successfully registered!!!"})
	return nil
}

// Add Product with Image Upload
func (r *Repository) AddProduct(context *fiber.Ctx) error {
	product := Product{}
	err := context.BodyParser(&product)
	if err != nil {
		// Handle parsing error
		return err
	}

	// Handle image upload
	file, err := context.FormFile("image")
	if err != nil {
		// Handle image upload error
		return err
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		// Handle file open error
		return err
	}
	defer src.Close()

	// Read the file data into a byte slice
	imageData, err := ioutil.ReadAll(src)
	if err != nil {
		// Handle read error
		return err
	}

	// Store the image data in the product object
	product.ImageData = imageData

	// Insert the product (including image data) into the database
	if err := r.DB.Table("product").Create(&product).Error; err != nil {
		// Handle database insert error
		return err
	}

	// Return a success response
	return context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Product added successfully"})
}

// Add Rider Account
func (r *Repository) AddRiderAccount(context *fiber.Ctx) error {
	riderAccount := RiderAccount{}
	err := context.BodyParser(&riderAccount)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "invalid request"})
		return err
	}

	if riderAccount.Password != riderAccount.Confirm_Password {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "passwords do not match"})
		return nil
	}

	// Additional validation and logic can be added here

	// Hash the password
	hashedPassword, err := hashPassword(riderAccount.Password)
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "error hashing password"})
		return err
	}

	// Create the new rider account
	newRiderAccount := RiderAccount{
		Fullname:         riderAccount.Fullname,
		Age:              riderAccount.Age,
		Address:          riderAccount.Address,
		Email:            riderAccount.Email,
		Username:         riderAccount.Username,
		DriverLicenseNum: riderAccount.DriverLicenseNum,
		Password:         hashedPassword,
	}

	// Store the rider account in the database (you will need to define the RiderAccount model and database table)
	err = r.DB.Table("rider_account").Create(&newRiderAccount).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not create rider account"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Rider Account added successfully"})
	return nil
}

// Handle purchase submission
func (r *Repository) SubmitPurchase(context *fiber.Ctx) error {
	purchase := Order{}
	err := context.BodyParser(&purchase)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "Invalid request"})
		return err
	}

	// Store the purchase in the database
	err = r.DB.Table("orders").Create(&purchase).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "Could not create purchase"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Purchase saved successfully"})
	return nil
}

// log in
func (r *Repository) Login(context *fiber.Ctx) error {
	loginRequest := LoginRequest{}
	Clientrespones := Account{}

	err := context.BodyParser(&loginRequest)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "invalid request"})
		return err
	}

	err = r.DB.Table("account").Where("username = ?", loginRequest.Username).First(&Clientrespones).Error
	if err != nil {
		context.Status(http.StatusUnauthorized).JSON(
			&fiber.Map{"message": "Invalid Username or Password"})
		return nil
	}

	// Check if the provided password matches the hashed password in the database
	err = bcrypt.CompareHashAndPassword([]byte(Clientrespones.Password), []byte(loginRequest.Password))
	if err != nil {
		context.Status(http.StatusUnauthorized).JSON(
			&fiber.Map{"message": "Invalid Username or Password"})
		return nil
	}

	textMessage := Message{}
	textMessage.Message = "Welcome! " + loginRequest.Username
	return context.JSON(textMessage)
}

// Update user account
func (r *Repository) UpdateAccount(context *fiber.Ctx) error {
	var updateRequest UpdateAccountRequest
	if err := context.BodyParser(&updateRequest); err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "Invalid request"})
		return err
	}

	// Update the user's account details in the database based on the username
	err := r.DB.Table("account").
		Where("username = ?", updateRequest.Username).
		Updates(&Account{
			Email: updateRequest.Email,
		}).Error

	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to update account"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Account updated successfully"})
	return nil
}

// Update user account by Admin
func (r *Repository) UpdateUser(context *fiber.Ctx) error {
	var updateRequest UpdateUserRequest
	if err := context.BodyParser(&updateRequest); err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "Invalid request"})
		return err
	}

	// Update the user's account details in the database based on the username
	err := r.DB.Table("account").
		Where("username = ?", updateRequest.Username).
		Updates(&Account{
			Fullname: updateRequest.Fullname,
			Email:    updateRequest.Email,
		}).Error

	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to update user"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "User updated successfully"})
	return nil
}

// Update Product by Admin
func (r *Repository) UpdateProductByTitle(context *fiber.Ctx) error {
	title := context.Query("title")

	// Check if the product exists
	var existingProduct Product
	err := r.DB.Table("product").
		Where("title = ?", title).
		First(&existingProduct).Error

	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "Product not found"})
		return err
	}

	// Parse the updated product data from the request body
	var updatedProduct Product
	if err := context.BodyParser(&updatedProduct); err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "Invalid request"})
		return err
	}

	// Update the product in the database
	err = r.DB.Table("product").
		Where("title = ?", title).
		Updates(&Product{
			Title:       updatedProduct.Title,
			Description: updatedProduct.Description,
			Price:       updatedProduct.Price,
			Quantity:    updatedProduct.Quantity,
			ImageData:   updatedProduct.ImageData,
		}).Error

	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to update product"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Product updated successfully"})
	return nil
}

// Update Rider account by Admin
func (r *Repository) UpdateRiderAccountByUsername(context *fiber.Ctx) error {
	var updateRequest UpdateRiderAccountRequest
	if err := context.BodyParser(&updateRequest); err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "Invalid request"})
		return err
	}

	var existingRiderAccount RiderAccount
	err := r.DB.Table("rider_account").
		Where("username = ?", updateRequest.Username).
		First(&existingRiderAccount).Error

	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "Rider account not found"})
		return err
	}

	err = r.DB.Table("rider_account").
		Where("username = ?", updateRequest.Username).
		Updates(&RiderAccount{
			Fullname:         updateRequest.Fullname,
			Age:              updateRequest.Age,
			Address:          updateRequest.Address,
			Email:            updateRequest.Email,
			DriverLicenseNum: updateRequest.DriverLicenseNum,
		}).Error

	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to update rider account"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Rider account updated successfully"})
	return nil
}

// Change password
func (r *Repository) UpdatePassword(context *fiber.Ctx) error {
	var updateRequest UpdatePasswordRequest
	if err := context.BodyParser(&updateRequest); err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "Invalid request"})
		return err
	}

	var existingAccount Account
	err := r.DB.Table("account").
		Where("username = ?", updateRequest.Username).
		First(&existingAccount).Error

	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "User not found"})
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingAccount.Password), []byte(updateRequest.CurrentPassword))
	if err != nil {
		context.Status(http.StatusUnauthorized).JSON(
			&fiber.Map{"message": "Invalid current password"})
		return nil
	}

	// Hash the new password
	hashedPassword, err := hashPassword(updateRequest.NewPassword)
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Error hashing new password"})
		return err
	}

	// Update the user's password in the database
	err = r.DB.Table("account").
		Where("username = ?", updateRequest.Username).
		Update("password", hashedPassword).Error

	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to update password"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Password updated successfully"})
	return nil
}

// Get fullname & email by username
func (r *Repository) GetUserData(context *fiber.Ctx) error {
	username := context.Query("username")

	var userData struct {
		Fullname string `json:"full_name"`
		Email    string `json:"email"`
		Address  string `json:"address"`
	}

	err := r.DB.Table("account").
		Select("fullname, email, address").
		Where("username = ?", username).
		First(&userData).Error

	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "User not found"})
		return err
	}

	return context.JSON(userData)
}

// GetUserData by username
func (r *Repository) GetUserData2(context *fiber.Ctx) error {
	username := context.Query("username")

	var userData struct {
		Fullname string `json:"fullname"`
		Age      int    `json:"age"`
		Address  string `json:"address"`
		Email    string `json:"email"`
	}

	err := r.DB.Table("account").
		Select("fullname, age, address, email").
		Where("username = ?", username).
		First(&userData).Error

	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "User not found"})
		return err
	}

	return context.JSON(userData)
}

// Get all user accounts
func (r *Repository) GetAllAccounts(context *fiber.Ctx) error {
	var accounts []Account

	// Retrieve all user accounts from the database
	err := r.DB.Table("account").Find(&accounts).Error
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to retrieve user accounts"})
		return err
	}

	return context.JSON(accounts)
}

// Get all usernames
func (r *Repository) GetAllUsernames(context *fiber.Ctx) error {
	var usernames []string

	// Retrieve all usernames from the database
	err := r.DB.Table("account").Pluck("username", &usernames).Error
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to retrieve usernames"})
		return err
	}

	return context.JSON(usernames)
}

// Get all rider usernames
func (r *Repository) GetAllRiderUsernames(context *fiber.Ctx) error {
	var usernames []string

	// Retrieve all rider usernames from the rider_account table
	err := r.DB.Table("rider_account").Pluck("username", &usernames).Error
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to retrieve rider usernames"})
		return err
	}

	return context.JSON(usernames)
}

// Get all products
func (r *Repository) GetAllProducts(context *fiber.Ctx) error {
	var products []Product

	// Retrieve all products from the database
	err := r.DB.Table("product").Find(&products).Error
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to retrieve products"})
		return err
	}

	return context.JSON(products)
}

// Get all Products Titles
func (r *Repository) GetAllProductTitles(context *fiber.Ctx) error {
	var productTitles []string

	// Retrieve all product titles from the database
	err := r.DB.Table("product").Pluck("title", &productTitles).Error
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to retrieve product titles"})
		return err
	}

	return context.JSON(productTitles)
}

// Get all loan applicant full names
func (r *Repository) GetAllLoanApplicantFullNames(context *fiber.Ctx) error {
	var fullNames []string

	// Retrieve all loan applicants from the database
	err := r.DB.Table("loanapplication").Pluck("full_name", &fullNames).Error
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to retrieve loan applicant full names"})
		return err
	}

	return context.JSON(fullNames)
}

// Get LoanApplicant info by Admin
func (r *Repository) GetLoanApplicantByFullname(context *fiber.Ctx) error {
	fullname := context.Query("fullname") // Get the fullname from query parameters

	var loanApplicant LoanApplication

	// Retrieve loan applicant information from the database based on the fullname
	err := r.DB.Table("loanapplication").
		Where("full_name = ?", fullname).
		First(&loanApplicant).Error

	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "Loan applicant not found"})
		return err
	}

	return context.JSON(loanApplicant)
}

// Get CartItems

// Delete user account by Admin
func (r *Repository) DeleteAccount(context *fiber.Ctx) error {
	username := context.Query("username")

	var existingAccount Account
	err := r.DB.Table("account").
		Where("username = ?", username).
		First(&existingAccount).Error

	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "User not found"})
		return err
	}

	err = r.DB.Table("account").
		Where("username = ?", username).
		Delete(&Account{}).Error

	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to delete user account"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "User account deleted successfully"})
	return nil
}

// Deletes a product by Admin
func (r *Repository) DeleteProduct(context *fiber.Ctx) error {
	// Get the product title from the query parameters
	title := context.Query("title")

	// Check if the product exists
	var existingProduct Product
	err := r.DB.Table("product").
		Where("title = ?", title).
		First(&existingProduct).Error

	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "Product not found"})
		return err
	}

	// Delete the product from the database
	err = r.DB.Table("product").
		Where("title = ?", title).
		Delete(&Product{}).Error

	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to delete product"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Product deleted successfully"})
	return nil
}

// add product to cart
func (r *Repository) AddToCart(ctx *fiber.Ctx) error {

	item := CartItem{}

	if err := ctx.BodyParser(&item); err != nil {
		ctx.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
			"message": "Request failed",
		})
		return err
	}

	r.CartMap[item.ProductID] += item.Quantity

	ctx.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Product added to cart successfully",
		"data":    r.CartMap,
	})
	return nil
}

// remove product from the cart
func (r *Repository) RemoveFromCart(ctx *fiber.Ctx) error {
	productIDStr := ctx.Params("product_id")

	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		ctx.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
			"message": "Invalid product ID",
		})
		return err
	}

	delete(r.CartMap, uint(productID))

	ctx.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Product removed from cart successfully",
		"data":    r.CartMap,
	})
	return nil
}

// Routes
func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// Log In
	api.Post("/login", r.Login)
	// Create & Add
	api.Post("/create_account", r.CreateAccount)
	api.Post("/create_loan_application", r.CreateLoanApplication)
	api.Post("/add_product", r.AddProduct)
	api.Post("/add_rider_account", r.AddRiderAccount)
	api.Post("/submit_purchase", r.SubmitPurchase)

	// Update
	api.Put("/update_account", r.UpdateAccount)
	api.Put("/update_password", r.UpdatePassword)
	api.Put("/update_user", r.UpdateUser)
	api.Put("/update_product_by_title", r.UpdateProductByTitle)
	// Get
	api.Get("/get_user_data", r.GetUserData)
	api.Get("/get_userdata", r.GetUserData2)
	api.Get("/get_all_accounts", r.GetAllAccounts)
	api.Get("/get_all_usernames", r.GetAllUsernames)
	api.Get("/get_all_products", r.GetAllProducts)
	api.Get("/get_all_product_titles", r.GetAllProductTitles)

	//Delete
	api.Delete("/delete_account", r.DeleteAccount)
	api.Delete("/delete_product", r.DeleteProduct)

	api.Post("/add_to_cart", r.AddToCart)
	api.Post("/remove_from_cart/product_id", r.RemoveFromCart)
}

// .env
func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SLLMODE"),
	}

	db, err := storage.NewConnection(config)

	if err != nil {
		log.Fatal("could not load the database")
	}
	err = db.AutoMigrate(&LoanApplication{}, &Product{}, &Order{}, CartItem{})
	if err != nil {
		log.Fatal("could not auto-migrate database schema")
	}

	r := Repository{
		DB: db,
	}
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))
	r.SetupRoutes(app)
	app.Listen(":8080")
}
