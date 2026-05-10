package main

import (
	"os"
)

// @title           Module Library
// @version         1.0.0
// @termsOfService  https://github.com/BevisDev/godev

// @contact.name   Truong Thanh Binh
// @contact.url    https://github.com/BevisDev
// @contact.email  dev.binhtt@gmail.com
func main() {

	// 4. Tạo dữ liệu thực tế
	user1 := User{Name: "Anh Tuấn", Role: "Admin", Active: true}
	user2 := User{Name: "Bảo Nam", Role: "Member", Active: false}

	// 5. Render (Thực thi) template ra màn hình (os.Stdout)
	tmpl.Execute(os.Stdout, user1)
	tmpl.Execute(os.Stdout, user2)
}

type User struct {
	Name   string
	Role   string
	Active bool
}
