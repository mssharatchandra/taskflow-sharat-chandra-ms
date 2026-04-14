package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sharatchandra/taskflow-sharat-chandra-ms/internal/model"
)

func parsePagination(c *gin.Context) (model.PaginationParams, map[string]string) {
	params := model.PaginationParams{
		Page:  model.DefaultPage,
		Limit: model.DefaultLimit,
	}
	errs := make(map[string]string)

	if pageStr := strings.TrimSpace(c.Query("page")); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			errs["page"] = "must be a positive integer"
		} else {
			params.Page = page
		}
	}

	if limitStr := strings.TrimSpace(c.Query("limit")); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			errs["limit"] = "must be a positive integer"
		} else if limit > model.MaxLimit {
			errs["limit"] = "must be less than or equal to 100"
		} else {
			params.Limit = limit
		}
	}

	return params, errs
}
