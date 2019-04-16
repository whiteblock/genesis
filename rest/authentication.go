package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	util "../util"
	"crypto/x509"
	"encoding/pem"
	"github.com/Whiteblock/jwt-go"
)

const allowNoKeyAcess = true

func GetKey(kid string) (map[string]string, error) {
	res, err := util.HttpRequest("GET", "https://api.whiteblock.io/public/jwt-keys", "")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var keys []map[string]string
	err = json.Unmarshal([]byte(res), &keys)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for i := 0; i < len(keys); i++ {
		if keys[i]["kid"] == kid {
			return keys[i], nil
		}
	}
	return nil, fmt.Errorf("Could not find a matching entry for the kid")
}

func AuthN(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		next.ServeHTTP(w, r) //bypass
		return

		tokenString := r.Header.Get("Authorization")

		if len(tokenString) == 0 {
			log.Println("Info: Request came in without the Authorization header set")
			if allowNoKeyAcess {
				log.Println("Warning: Allowed access to request without a key")
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, "Missing JWT in Authorization header", 403)
			}
			return
		}
		tokenString = strings.Split(tokenString, " ")[1]
		//log.Printf("Given token is %s\n",tokenString)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			//Validate the key
			alg, ok := token.Header["alg"].(string)
			if !ok {
				return nil, fmt.Errorf("Malformed key, missing or invalid field \"alg\" in JWT header")
			}

			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("Malformed key, missing or invalid field \"kid\" in JWT header")
			}

			keyData, err := GetKey(kid)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if alg != keyData["alg"] {
				return nil, fmt.Errorf("Unexpected signing method: %s. Expected %s.", alg, keyData["alg"])
			}
			block, remaining := pem.Decode([]byte(keyData["pem"]))
			if block == nil {
				fmt.Printf("Remaining: %s", string(remaining))
				return nil, fmt.Errorf("Pem block is nil")
			}

			pub, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			return pub, nil
		})
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 403)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			fmt.Printf("%v\n", claims)
		} else {
			log.Println("Unknown claims type")
		}
		fmt.Printf("Token: %v\n", token)
		// Authenticated, move on to next step
		next.ServeHTTP(w, r)

	})
}
