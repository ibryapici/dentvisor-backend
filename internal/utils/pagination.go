package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PaginatedResponse, sayfalanmış API yanıtları için standart bir yapı sağlar.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// GetPaginationParams, istekteki page, limit ve search parametrelerini alır.
func GetPaginationParams(c *gin.Context) (page int, limit int, search string) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))
	search = c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	return page, limit, search
}

// Paginate, bir Gorm sorgusunu sayfalandırır ve PaginatedResponse döndürür.
func Paginate(db *gorm.DB, page, limit int, data interface{}) (PaginatedResponse, error) {
	var total int64
	// Toplam kayıt sayısını al
	if err := db.Count(&total).Error; err != nil {
		return PaginatedResponse{}, err
	}

	// Sayfalandırma yap ve veriyi çek
	offset := (page - 1) * limit
	if err := db.Offset(offset).Limit(limit).Find(data).Error; err != nil {
		return PaginatedResponse{}, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}
