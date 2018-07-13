/*
* Sample API with GET and POST endpoint.
* POST data is converted to string and saved in internal memory.
* GET endpoint returns all strings in an array.
 */
package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func PKCS5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func PKCS5UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}

func tokenGenerator() string {
	b := make([]byte, 6)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

//KMS
type KeyData struct {
	Id  int    `json:"id"`
	Key string `json:"key"`
}

//eVault
type VaultData struct {
	Id          int    `json:"id"`
	Value       string `json:value`
	PartitionId string `json:partition_id`
}

func storeVault(val string) string {
	var username string = ""
	var passwd string = ""
	client := &http.Client{}
	q := "https://evaultdev.epointserver.com/api/v1/datastore/store_entry?program_id=1&partition_id=1&value=" + val
	req, err := http.NewRequest("POST", q, nil)
	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	s := string(bodyText)

	fmt.Printf(s)
	var res VaultData
	err = json.Unmarshal(bodyText, &res)

	if err != nil {
		fmt.Printf("Failed to unmarshal res %s, the error is %v", string(bodyText), err)

	}
	//fmt.Printf("key %v", res.Key)
	//fmt.Printf(res.Key)
	i := fmt.Sprintf("%v", res.Id)
	return i
}

func getKMS() (string, string) {
	var username string = ""
	var passwd string = ""
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://kmsdev.epointserver.com/api/v1/key/generate_key?program_id=1&length=24", nil)
	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	s := string(bodyText)

	fmt.Printf(s)
	var res KeyData
	err = json.Unmarshal(bodyText, &res)

	if err != nil {
		fmt.Printf("Failed to unmarshal res %s, the error is %v", string(bodyText), err)

	}

	i := fmt.Sprintf("%v", res.Id)
	return i, res.Key
}

func showKey(key_id string) string {
	var username string = ""
	var passwd string = ""
	client := &http.Client{}
	u := fmt.Sprintf("https://kmsdev.epointserver.com/api/v1/key/get_key?program_id=1&id=%s", key_id)
	req, err := http.NewRequest("GET", u, nil)
	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	s := string(bodyText)

	fmt.Printf(s)
	var res KeyData
	err = json.Unmarshal(bodyText, &res)

	if err != nil {
		fmt.Printf("Failed to unmarshal res %s, the error is %v", string(bodyText), err)

	}

	//i := fmt.Sprintf("%v", res.Id)
	return res.Key
}

func decryptData(en []byte, key string, l int) string {

	block, err := des.NewTripleDESCipher([]byte(key))

	if err != nil {
		fmt.Printf("%s \n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("%d bytes NewTripleDESCipher key with block size of %d bytes\n", len(key), block.BlockSize)

	ciphertext := []byte("abcdef1234567890")
	iv := ciphertext[:des.BlockSize] // const BlockSize = 8

	//decrypt

	decrypter := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, l)
	decrypter.CryptBlocks(decrypted, en)

	decrypted = PKCS5UnPadding(decrypted)

	fmt.Printf("This is the second %x decrypt to %s\n", en, decrypted)

	return string(decrypted)

}

func encryptData(val string) (string, string, string) {
	//get key
	key_id, triplekey := getKMS()
	fmt.Println("get KMS")
	fmt.Println(key_id)
	fmt.Println(triplekey)
	fmt.Println("-------")

	plaintext := []byte(val)
	block, err := des.NewTripleDESCipher([]byte(triplekey))

	if err != nil {
		fmt.Printf("%s \n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("%d bytes NewTripleDESCipher key with block size of %d bytes\n", len(triplekey), block.BlockSize)

	ciphertext := []byte("abcdef1234567890")
	iv := ciphertext[:des.BlockSize] // const BlockSize = 8

	// encrypt
	mode := cipher.NewCBCEncrypter(block, iv)

	plaintext = PKCS5Padding(plaintext, block.BlockSize())

	encrypted := make([]byte, len(plaintext))
	mode.CryptBlocks(encrypted, plaintext)
	fmt.Printf("key %s \n", triplekey)
	fmt.Printf("%s encrypt to %x \n", plaintext, encrypted)
	n := fmt.Sprintf("%x", encrypted)
	//n := string(encrypted)
	fmt.Println("here is the encrypted data in bytes")
	fmt.Println(encrypted)
	fmt.Println("here is the string")
	fmt.Println(string(encrypted[:]))
	fmt.Println("here is the hex value")
	str := hex.EncodeToString(encrypted)
	fmt.Println(str)

	//hex to bytes
	//fmt.Println("hex to bytes")
	//src := n
	//bs, err := hex.DecodeString(src)
	//fmt.Println(string(bs))
	//fmt.Println(src)
	//fmt.Println("here is the hex to bytes result")
	//fmt.Println(bs)

	//decrypt
	//decrypter := cipher.NewCBCDecrypter(block, iv)
	//decrypted := make([]byte, len(plaintext))
	//decrypter.CryptBlocks(decrypted, encrypted)

	//decrypted = PKCS5UnPadding(decrypted)

	//fmt.Printf("%x decrypt to %s and length %s \n", encrypted, decrypted, len(plaintext))
	//e := []byte("08deb121f30b4d29a44f2b0179a0c8697bf263a3e3621f3e")
	//k := "1244d54e26f38129e8eca527"
	//o := "ryan.amog@gmail.com"
	//decryptData(bs, triplekey, len(plaintext))
	sl := fmt.Sprintf("%v", len(plaintext))
	return key_id, n, sl
}

// nested within sbserver response
type Details struct {
	Token          string
	ItemType       string
	CustomerNumber string
}

// response from server
type Results struct {
	Data Details
}

type VaultDetails struct {
	ItemType string `json.ItemType`
	TextItem string `json.TextItem`
}

type VaultItem struct {
	VaultItem VaultDetails
}

//get key
type KeyResults struct {
	Data VaultItem
}

var (
	// flagPort is the open port the application listens on
	flagPort = flag.String("port", "3000", "Port to listen on")
)

var results []string

// PostHandler converts post request body to string
func PostEncodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "Authorization Failed", http.StatusUnauthorized)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])

		pair := strings.SplitN(string(payload), ":", 2)

		auth_val, _ := base64.StdEncoding.DecodeString(pair[0])

		if string(auth_val) != "eclypsesgo:2018" {
			http.Error(w, string(auth_val), http.StatusUnauthorized)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}
		p := strings.Split(string(body), "&")
		//P01 = item_type
		//P02 = item_value

		var itype string = strings.Replace(p[0], "P01=", "", -1)
		var ival string = strings.Replace(p[1], "P02=", "", -1)

		//process encryption
		key_id, en, s_len := encryptData(ival)

		message := map[string]interface{}{
			"Header": map[string]string{
				"MerchantToken":    "",
				"SenderToken":      "",
				"AuthenticationId": "",
			},
			"StoreVaultRequest": map[string]string{
				"ItemType":       itype,
				"TextItem":       en,
				"CustomerNumber": "",
				"Name":           "hello world",
			},
		}

		fmt.Println("Here it is %x", en)

		bytesRepresentation, err := json.MarshalIndent(message, "", "    ")

		resp, err := http.Post("https://test-certainstore.transcertain.com/StoreVault", "application/json", bytes.NewBuffer(bytesRepresentation))
		if err != nil {
			//log.Fatal(err)
			//if erroro generate own token
			a := "LUN" + tokenGenerator()
			fmt.Println(a)

			respmessage := map[string]interface{}{
				"ResponseCode": "0000",
				"Token":        a,
			}

			sv := storeVault(en)

			db, err := sql.Open("mysql", "")
			if err != nil {
				panic(err.Error())
			}
			defer db.Close()

			// Execute the query
			rows, err := db.Query("INSERT INTO myVault (token, key_id, vault_id, string_len) VALUES (?, ?, ?, ?)", a, key_id, sv, s_len)
			if err != nil {
				panic(err.Error()) // proper error handling instead of panic in your app
			}
			defer rows.Close()

			bytesResponse, err := json.MarshalIndent(respmessage, "", "    ")

			fmt.Fprint(w, string(bytesResponse))
			return
		}

		respbody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var res Results
		//var response
		json.Unmarshal(respbody, &res)

		var token string = res.Data.Token
		//fmt.Printf("Results: %v\n", res)
		//fmt.Printf("Token: %s\n",res.Data.Token)
		respmessage := map[string]interface{}{
			"ResponseCode": "0000",
			"Token":        token,
		}

		sv := storeVault(en)

		db, err := sql.Open("mysql", "")
		if err != nil {
			panic(err.Error())
		}
		defer db.Close()

		// Execute the query
		rows, err := db.Query("INSERT INTO myVault (token, key_id, vault_id, string_len) VALUES (?, ?, ?, ?)", token, key_id, sv, s_len)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		defer rows.Close()

		bytesResponse, err := json.MarshalIndent(respmessage, "", "    ")

		//fmt.Fprint(w, token)
		//results = append(results, string(body))
		fmt.Fprint(w, string(bytesResponse))
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// PostHandler converts post request body to string
func PostDecodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "Authorization Failed", http.StatusUnauthorized)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])

		pair := strings.SplitN(string(payload), ":", 2)

		auth_val, _ := base64.StdEncoding.DecodeString(pair[0])

		if string(auth_val) != "eclypsesgo:2018" {
			http.Error(w, string(auth_val), http.StatusUnauthorized)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}
		p := strings.Split(string(body), "&")

		//P01 = token

		var token string = strings.Replace(p[0], "P01=", "", -1)

		message := map[string]interface{}{
			"Header": map[string]string{
				"MerchantToken":    "",
				"SenderToken":      "",
				"AuthenticationId": "",
			},
			"RetrieveVaultRequest": map[string]string{
				"Token": token,
			},
		}

		bytesRepresentation, err := json.MarshalIndent(message, "", "    ")

		resp, err := http.Post("https://test-certainstore.transcertain.com/RetrieveVault", "application/json", bytes.NewBuffer(bytesRepresentation))
		if err != nil {
			log.Fatal(err)
		}

		respbody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(respbody))
		var res KeyResults
		//var response
		json.Unmarshal(respbody, &res)

		var itemText string = res.Data.VaultItem.TextItem

		db, err := sql.Open("mysql", "")
		if err != nil {
			panic(err.Error())
		}
		defer db.Close()

		// Execute the query
		rows, err := db.Query("SELECT key_id FROM myVault WHERE token=?", token)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		var key_id string
		for rows.Next() {
			//var key_id string
			rows.Scan(&key_id)
			//fmt.Println("KEY ID ?", key_id)

		}
		defer rows.Close()

		kms := showKey(key_id)

		fmt.Println("KMS key ?", kms)

		bs, err := hex.DecodeString(itemText)
		decrypt_val := decryptData(bs, kms, 24)
		//fmt.Printf("Results: %v\n", res)
		//fmt.Printf("Token: %s\n",res.Data.Token)
		respmessage := map[string]interface{}{
			"ResponseCode": "0000",
			"VaultItem":    decrypt_val,
		}

		bytesResponse, err := json.MarshalIndent(respmessage, "", "    ")

		//fmt.Fprint(w, token)
		//results = append(results, string(body))
		//fmt.Fprint(w, string(respbody))
		fmt.Fprint(w, string(bytesResponse))
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	flag.Parse()
}

func main() {
	mux := http.NewServeMux()
	////mux.HandleFunc("/", GetHandler)
	mux.HandleFunc("/storeVault", PostEncodeHandler)
	mux.HandleFunc("/retrieveVault", PostDecodeHandler)

	log.Printf("listening on port %s", *flagPort)
	log.Fatal(http.ListenAndServe(":"+*flagPort, mux))
}
