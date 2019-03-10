package main

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

//secret key
var secretKey = []byte("abcd1234!@#$")

type ArithmeticCustomClaims struct {
	UserId string `json:"userId"`
	Name   string `json:"name"`

	jwt.StandardClaims
}

func jwtKeyFunc(token *jwt.Token) (interface{}, error) {
	return secretKey, nil
}

func Sign(name, uid string) (string, error) {

	//两分钟后过期
	expAt := time.Now().Add(time.Duration(2) * time.Minute).Unix()

	claims := ArithmeticCustomClaims{
		UserId: uid,
		Name:   name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expAt,
			Issuer:    "system",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
