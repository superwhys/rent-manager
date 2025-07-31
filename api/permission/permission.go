// File:		permission.go
// Created by:	Hoven
// Created on:	2025-07-31
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package permission

import "github.com/miebyte/goutils/ginutils"

type App interface {
}

func Router(app App) ginutils.EngineOption {
	return ginutils.WithGroupRouters("/permission")
}
