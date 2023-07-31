package main

import (
	f "fmt"

	"errors"

	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// album represents data about a record album.

type input struct {
	Folio_No        string  `json:"Folio_No" gorm:"primaryKey"`
	Description     string  `json:"Description"`
	Unit            string  `json:"Unit"`
	Stock_Qty       int     `json:"Stock_Qty"`
	WAC             float64 `json:"WAC"`
	Bin_Location    string  `json:"Bin_Location"`
	Remarks         string  `json:"Remarks"`
	Shelf_life_item string  `json:"Shelf_life_item"`
}

type submit struct {
	New    []input `json:"new"`
	Change []input `json:"change"`
}
type loginInfo struct {
	Username string `json:"username" gorm:"primaryKey"`
	Password string `json:"password"`
}
type UserInfo struct {
	Username string `json:"username" gorm:"primaryKey"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
type filter struct {
	Description string `json:"description"`
	Remarks     string `json:"remarks"`
}

func main() {

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowWildcard:   true,
		AllowMethods:    []string{"*"},
		AllowHeaders:    []string{"*"},
	}))

	router.POST("/updateDB", updateDB)
	router.POST("/submitDB", submitDB)
	router.POST("/getDB", getDB)
	router.POST("/delete", delete)
	router.POST("/add", add)
	router.POST("/login", login)
	router.Run("localhost:8080")
}

func getDB(c *gin.Context) {
	var receiveFilter filter

	dsn := "root:Password123!@tcp(localhost:3306)/Stock_balance"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		f.Println("error validating gorm.Open")
		panic(err.Error())
	}
	if err := c.BindJSON(&receiveFilter); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error)
		f.Println(err.Error())
		return
	}
	var description = receiveFilter.Description
	var remarks = receiveFilter.Remarks
	if description == "" && remarks == "" {
		var receiveData []input
		result := db.Table("inputs").Find(&receiveData)
		if result.Error != nil {
			c.IndentedJSON(http.StatusNotFound, result.Error)
			f.Println("read error")
		}
		c.IndentedJSON(http.StatusAccepted, receiveData)
	} else if description != "" && remarks == "" {
		var receiveData []input
		result := db.Table("inputs").Where("description LIKE @value", map[string]interface{}{"value": "%" + description + "%"}).Find(&receiveData)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.IndentedJSON(http.StatusAccepted, "empty")
				return
			}
		}
		c.IndentedJSON(http.StatusAccepted, receiveData)
	} else if description == "" && remarks != "" {
		var receiveData []input
		result := db.Table("inputs").Where("remarks LIKE @value", map[string]interface{}{"value": "%" + remarks + "%"}).Find(&receiveData)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.IndentedJSON(http.StatusAccepted, "empty")
				return
			}
		}
		c.IndentedJSON(http.StatusAccepted, receiveData)
	} else if description != "" && remarks != "" {
		var receiveData []input
		result := db.Table("inputs").Where("description LIKE @value1 && remarks LIKE @value2", map[string]interface{}{"value1": "%" + description + "%"}, map[string]interface{}{"value2": "%" + remarks + "%"}).Find(&receiveData)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.IndentedJSON(http.StatusAccepted, "empty")
				return
			}
		}
		c.IndentedJSON(http.StatusAccepted, receiveData)
	}

}

func login(c *gin.Context) {
	var login loginInfo
	dsn := "root:Password123!@tcp(localhost:3306)/Stock_balance"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		f.Println("error validating gorm.Open")
		panic(err.Error())
	}
	if err := c.BindJSON(&login); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error)
		f.Println(err.Error())
		return
	}
	var receiveData UserInfo
	var username = login.Username
	var password = login.Password
	result := db.Table("user").Where("username = ? AND password = ? ", username, password).First(&receiveData)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.IndentedJSON(http.StatusAccepted, "invalid")
			return
		}

	}
	c.IndentedJSON(http.StatusAccepted, gin.H{"role": receiveData.Role, "username": username})
}

func add(c *gin.Context) {
	var dbdata input

	dsn := "root:Password123!@tcp(localhost:3306)/Stock_balance"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		f.Println("error validating gorm.Open")
		panic(err.Error())
	}

	if err := c.BindJSON(&dbdata); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error)
		f.Println(err.Error())
		return
	}

	db.AutoMigrate(&dbdata)

	var receiveData []input
	result := db.Table("inputs").Find(&receiveData)
	if result.Error != nil {
		c.IndentedJSON(http.StatusNotFound, result.Error)
		f.Println("read error")
	}

	var sendData input
	var repeated bool
	repeated = false
	for k := 0; k < len(receiveData); k++ {
		if dbdata.Folio_No == receiveData[k].Folio_No {
			repeated = true
			sendData = receiveData[k]
			break

		}
	}
	if repeated == true {
		c.IndentedJSON(http.StatusCreated, sendData)
	} else {
		result := db.Table("inputs").Create(dbdata)
		if result.Error != nil {
			f.Println(result.Error)
			panic(result.Error)
		}
	}

}

func updateDB(c *gin.Context) {
	var dbdata []input

	dsn := "root:Password123!@tcp(localhost:3306)/Stock_balance"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		f.Println("error validating gorm.Open")
		panic(err.Error())
	}

	if err := c.BindJSON(&dbdata); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error)
		f.Println(err.Error())
		return
	}

	db.AutoMigrate(&dbdata)

	var duplicated []input
	var new []input

	var receiveData []input
	result := db.Table("inputs").Find(&receiveData)
	if result.Error != nil {
		c.IndentedJSON(http.StatusNotFound, result.Error)
		f.Println("read error")
	}

	for i := 0; i < len(dbdata); i++ {
		var repeated bool
		repeated = false
		for k := 0; k < len(receiveData); k++ {
			if dbdata[i].Folio_No == receiveData[k].Folio_No {
				repeated = true
				duplicated = append(duplicated, dbdata[i])

			}
		}
		if repeated == false {
			new = append(new, dbdata[i])
		}
	}
	//var combined = ["new": new, "change": duplicated]
	c.IndentedJSON(http.StatusCreated, gin.H{"new": new, "change": duplicated})

	/*results := db.Table("inputs").Create(dbdata)
	if results.Error != nil {
		f.Println("Error")
		panic(results.Error)
	}

	// Add the new album to the slice.
	c.IndentedJSON(http.StatusCreated, receiveData)*/

}

func delete(c *gin.Context) {
	var dbdata input

	dsn := "root:Password123!@tcp(localhost:3306)/Stock_balance"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		f.Println("error validating gorm.Open")
		panic(err.Error())
	}

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&dbdata); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error)
		f.Println(err.Error())
		return
	}

	result := db.Table("inputs").Delete(dbdata)
	if result.Error != nil {
		f.Println("Delete Error")
		panic(result.Error)
	}
}

func submitDB(c *gin.Context) {
	var dbdata submit

	dsn := "root:Password123!@tcp(localhost:3306)/Stock_balance"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		f.Println("error validating gorm.Open")
		panic(err.Error())
	}

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&dbdata); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error)
		f.Println(err.Error())
		return
	}

	var duplicated []input
	var new []input

	new = dbdata.New
	duplicated = dbdata.Change

	//var combined = ["new": new, "change": duplicated]
	if new != nil {
		results := db.Table("inputs").Create(new)
		if results.Error != nil {
			f.Println("Error")
			panic(results.Error)
		}
	}

	if duplicated != nil {
		result := db.Table("inputs").Save(duplicated)
		if result.Error != nil {
			f.Println("Error")
			panic(result.Error)
		}
	}

	// Add the new album to the slice.
	c.IndentedJSON(http.StatusCreated, new)
}
