package repository

import (
	"ecommerce-backend/internal/model"

	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建商品仓储
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *model.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) GetByID(id uint) (*model.Product, error) {
	var product model.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetByIDWithDetails(id uint) (*model.Product, error) {
	var product model.Product
	err := r.db.
		Preload("Category").
		Preload("Images").
		Preload("Tags").
		Preload("Reviews").
		First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetBySlug(slug string) (*model.Product, error) {
	var product model.Product
	err := r.db.
		Preload("Category").
		Preload("Images").
		Preload("Tags").
		Where("slug = ?", slug).
		First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetBySKU(sku string) (*model.Product, error) {
	var product model.Product
	err := r.db.Where("sku = ?", sku).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) Update(product *model.Product) error {
	return r.db.Save(product).Error
}

func (r *productRepository) Delete(id uint) error {
	return r.db.Delete(&model.Product{}, id).Error
}

func (r *productRepository) List(params ProductListParams) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	query := r.db.Model(&model.Product{}).Preload("Category").Preload("Images")

	// 应用过滤条件
	if params.CategoryID != nil {
		query = query.Where("category_id = ?", *params.CategoryID)
	}

	if params.Brand != "" {
		query = query.Where("brand ILIKE ?", "%"+params.Brand+"%")
	}

	if params.MinPrice != nil {
		query = query.Where("price >= ?", *params.MinPrice)
	}

	if params.MaxPrice != nil {
		query = query.Where("price <= ?", *params.MaxPrice)
	}

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序
	orderBy := "created_at DESC"
	if params.SortBy != "" {
		order := "DESC"
		if params.Order == "asc" {
			order = "ASC"
		}
		switch params.SortBy {
		case "price":
			orderBy = "price " + order
		case "name":
			orderBy = "name " + order
		case "rating":
			orderBy = "rating " + order
		case "sales":
			orderBy = "sales_count " + order
		default:
			orderBy = "created_at " + order
		}
	}

	err := query.Order(orderBy).Offset(params.Offset).Limit(params.Limit).Find(&products).Error
	return products, total, err
}

func (r *productRepository) Search(query string, limit, offset int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	searchQuery := r.db.Model(&model.Product{}).
		Preload("Category").
		Preload("Images").
		Where("name ILIKE ? OR description ILIKE ? OR brand ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%")

	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := searchQuery.Offset(offset).Limit(limit).Find(&products).Error
	return products, total, err
}

func (r *productRepository) GetByCategoryID(categoryID uint, limit, offset int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	query := r.db.Model(&model.Product{}).
		Preload("Category").
		Preload("Images").
		Where("category_id = ? AND status = ?", categoryID, model.ProductStatusActive)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&products).Error
	return products, total, err
}

func (r *productRepository) GetFeaturedProducts(limit int) ([]model.Product, error) {
	var products []model.Product
	err := r.db.
		Preload("Category").
		Preload("Images").
		Where("status = ?", model.ProductStatusActive).
		Order("rating DESC, sales_count DESC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

func (r *productRepository) GetRelatedProducts(productID uint, limit int) ([]model.Product, error) {
	var products []model.Product

	// 获取当前商品的分类
	var currentProduct model.Product
	if err := r.db.First(&currentProduct, productID).Error; err != nil {
		return products, err
	}

	err := r.db.
		Preload("Category").
		Preload("Images").
		Where("category_id = ? AND id != ? AND status = ?",
			currentProduct.CategoryID, productID, model.ProductStatusActive).
		Order("rating DESC").
		Limit(limit).
		Find(&products).Error

	return products, err
}

func (r *productRepository) UpdateStock(id uint, quantity int) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).
		Update("stock", gorm.Expr("stock - ?", quantity)).Error
}

func (r *productRepository) UpdateViewCount(id uint) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).
		Update("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *productRepository) UpdateRating(productID uint) error {
	// 重新计算平均评分和评价数量
	var avgRating float64
	var reviewCount int64

	r.db.Model(&model.Review{}).
		Where("product_id = ?", productID).
		Count(&reviewCount)

	if reviewCount > 0 {
		r.db.Model(&model.Review{}).
			Where("product_id = ?", productID).
			Select("AVG(rating)").
			Scan(&avgRating)
	}

	return r.db.Model(&model.Product{}).Where("id = ?", productID).
		Updates(map[string]interface{}{
			"rating":       avgRating,
			"review_count": reviewCount,
		}).Error
}

func (r *productRepository) GetLowStockProducts(threshold int) ([]model.Product, error) {
	var products []model.Product
	err := r.db.
		Where("stock <= ? AND track_stock = ? AND status = ?",
			threshold, true, model.ProductStatusActive).
		Find(&products).Error
	return products, err
}

func (r *productRepository) BulkUpdateStatus(ids []uint, status model.ProductStatus) error {
	return r.db.Model(&model.Product{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}
