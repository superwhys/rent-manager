package main

import (
	"github.com/superwhys/rent-manager-server/model"
	"gorm.io/gen"
)

//go:generate go run generate.go
func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:       "../../dal",
		WithUnitTest:  false,
		FieldNullable: true,
		Mode:          gen.WithQueryInterface,
	})

	g.ApplyBasic(
		&model.User{},
		&model.Room{},
		&model.Tenant{},
		&model.Property{},
		&model.Bill{},
		&model.MeterReading{},
		&model.Permission{},
		&model.SystemSettings{},
	)
	g.Execute()
}
