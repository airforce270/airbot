// Package model defines database data models.
package model

import (
	"gorm.io/gorm"
)

// AllModels contains one of each defined data model, for auto-migrations.
var AllModels = []gorm.Model{}
