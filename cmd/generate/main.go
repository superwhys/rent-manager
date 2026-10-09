package main

import (
	"gorm.io/gen"

	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
)

func main() {
	generator := gen.NewGenerator(gen.Config{
		OutPath: "../../internal/infra/persistence/mysql/query",
		Mode:    gen.WithQueryInterface,
	})
	generator.ApplyBasic(models.AllModels()...)
	generator.Execute()
}
