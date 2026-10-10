package main

//go:generate go run .

import (
	authmodels "github.com/miebyte/authkit/mysql/models"
	"gorm.io/gen"

	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
)

func main() {
	generator := gen.NewGenerator(gen.Config{
		OutPath: "../../internal/infra/persistence/mysql/query",
		Mode:    gen.WithQueryInterface,
	})
	generator.ApplyBasic(models.AllModels()...)
	generator.ApplyBasic(authmodels.Account{})
	generator.Execute()
}
