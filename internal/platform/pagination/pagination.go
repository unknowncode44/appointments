package pagination

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

type Page struct {
	Page     int32 `json:"page"`
	PageSize int32 `json:"page_size"`
	Offset   int32 `json:"-"`
}

type Response[T any] struct {
	Data     []T   `json:"data"`
	Page     int32 `json:"page"`
	PageSize int32 `json:"page_size"`
	Total    int64 `json:"total"`
}

func FromCtx(c *fiber.Ctx) Page {
	page := parsePositive(c.Query("page"), defaultPage)
	pageSize := parsePositive(c.Query("page_size"), defaultPageSize)
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return Page{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
	}
}

func parsePositive(raw string, fallback int32) int32 {
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return fallback
	}
	return int32(n)
}
