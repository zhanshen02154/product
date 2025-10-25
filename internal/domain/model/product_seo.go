package model

// ProductSeo
// @Description: 产品SEO模型
type ProductSeo struct {
	Id             int64  `gorm:"type:bigint(20);primaryKey;not null;autoIncrement" json:"id"`
	SeoTitle       string `gorm:"type:varchar(50);not null;default:''" json:"seo_title"`
	SeoKeyword    string `gorm:"type:varchar(100);not null;default:''" json:"seo_keyword"`
	SeoDescription string `gorm:"type:varchar(100);not null;default:''" json:"seo_description"`
	SeoProductId   int64  `gorm:"type:bigint(20);not null;default:0" json:"seo_product_id"`
}
