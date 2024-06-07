package controllers

import (
	"accounts_backend/database"
	"accounts_backend/models"
	request "accounts_backend/requests"
	"accounts_backend/response"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// SignUp function to handle the request
func SignUp(w http.ResponseWriter, r *http.Request) {
   //let's first validate our request
    var signUpReq request.SignUpRequest
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&signUpReq); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Initialize the validator and validate the request data
    validate := validator.New()
    err := validate.Struct(signUpReq)
    if err != nil {
        if errs, ok := err.(validator.ValidationErrors); ok {
			errMessages := request.CustomeErrorMessage(errs)
            http.Error(w, strings.Join(errMessages, ", "), http.StatusBadRequest)
            return
        }
        http.Error(w, "Validation failed", http.StatusInternalServerError)
        return
    }

    newUser := models.User{
        Username: signUpReq.Username,
        Name: signUpReq.Name,
        Email: signUpReq.Email,
		Phone: signUpReq.Phone,
		ImageProfile: signUpReq.ImageProfile,
		Role: signUpReq.Role,
        Password: signUpReq.Password,
        ID: uuid.NewString(), 
		Bio: signUpReq.Bio,
		CreatedAt: time.Now(), 
        UpdatedAt: time.Now(),
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Failed to hash password", http.StatusInternalServerError)
        return
    }
    newUser.Password = string(hashedPassword)
    //let have the logic to create the new users
    //lets first open the collection we need to insert the user in
    collection:= database.OpenCollection(database.Client,"users")
    ctx,cancel :=context.WithTimeout(context.Background(),10*time.Second)
    defer cancel()
    //let's now insert the user
    _,err=collection.InsertOne(ctx,newUser)
    if err != nil {

        err_response := response.ErrorResponse{
            OK: false,
            Message: err.Error(),
            Status: "failed",
        }
        json.NewEncoder(w).Encode(err_response)
    }


    // Properly respond with the created user, omitting sensitive data like password
    responseUser := newUser
    responseUser.Password = "" // Clear password before sending response
    w.Header().Set("Content-Type", "application/json")
    response:=response.Response{
        OK: true,
        Status: "success",
        Message: fmt.Sprintf("%s added successfully", newUser.Name),
        User: responseUser,
    }
    json.NewEncoder(w).Encode(response)

}


//let's have the signup function here
func Login(w http.ResponseWriter, r *http.Request) {
    var signInReq request.SignInRequest
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&signInReq); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Initialize the validator and validate the request data
    validate := validator.New()
    err := validate.Struct(signInReq)
    if err != nil {
        if errs, ok := err.(validator.ValidationErrors); ok {
        	errMessages := request.CustomeErrorMessage(errs)
            http.Error(w, strings.Join(errMessages, ", "), http.StatusBadRequest)
            return
        }
        http.Error(w, "Validation failed...", http.StatusInternalServerError)
        return
    }
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := database.OpenCollection(database.Client, "users")
       var user models.User
       err = collection.FindOne(ctx, bson.M{"username": signInReq.Username}).Decode(&user)
       if err != nil {
           if err == mongo.ErrNoDocuments {
            err_resp:=response.ErrorResponse{
                OK:false,
                Status:"failed",
                Message: "Invalid username or password" ,
            }
            errJSON, _ := json.Marshal(err_resp)
               http.Error(w, string(errJSON), http.StatusUnauthorized)
               return
           }
           http.Error(w, "Failed to find user", http.StatusInternalServerError)
           return
       }
   
       // Compare the password with the hashed password in the database
       err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(signInReq.Password))
       if err != nil {
        err_resp:=response.ErrorResponse{
            OK:false,
            Status:"failed",
            Message: "Invalid username or password" ,
        }
        errJSON, _ := json.Marshal(err_resp)
           http.Error(w, string(errJSON), http.StatusUnauthorized)
           return
       }
   
        token, err := generateJWT(user)
        if err != nil {
            http.Error(w, "Failed to generate token", http.StatusInternalServerError)
            return
        }
   
       // Return success response (including token if generated)
       w.Header().Set("Content-Type", "application/json")
	   user.Password = "" 
       json.NewEncoder(w).Encode(map[string]interface{}{
           "ok":     true,
           "status": "success",
           "message": "User signed in successfully",
           "users":user,
            "token":  token, 
       })
   
}

// func to generate JWT token
func generateJWT(user models.User) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour) 
    claims := &UserClaims{
        Email: user.Email,
		Roles: []string{user.Role},
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        return "", err
    }

    return tokenString, nil
}
var jwtKey = []byte("your_secret_key")

// UserClaims struct to hold custom claims for the token
type UserClaims struct {
    Email string `json:"email"`
	Roles []string `json:"roles"`
    jwt.RegisteredClaims
}
