package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"solemate/services/order-service/internal/domain/entity"
	"solemate/services/order-service/internal/domain/repository"
)

type orderRepositoryImpl struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) repository.OrderRepository {
	return &orderRepositoryImpl{
		db: db,
	}
}

func (r *orderRepositoryImpl) CreateOrder(ctx context.Context, order *entity.Order) error {
	// Generate order number if not provided
	if order.OrderNumber == "" {
		order.OrderNumber = r.generateOrderNumber()
	}

	// Calculate totals
	order.CalculateTotals()

	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepositoryImpl) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*entity.Order, error) {
	var order entity.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&order, "id = ?", orderID).Error

	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepositoryImpl) GetOrderByNumber(ctx context.Context, orderNumber string) (*entity.Order, error) {
	var order entity.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&order, "order_number = ?", orderNumber).Error

	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepositoryImpl) UpdateOrder(ctx context.Context, order *entity.Order) error {
	order.CalculateTotals()
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderRepositoryImpl) DeleteOrder(ctx context.Context, orderID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Order{}, "id = ?", orderID).Error
}

func (r *orderRepositoryImpl) GetOrdersByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Order, int64, error) {
	var orders []*entity.Order
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&entity.Order{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error

	return orders, total, err
}

func (r *orderRepositoryImpl) GetOrdersByStatus(ctx context.Context, status entity.OrderStatus, limit, offset int) ([]*entity.Order, int64, error) {
	var orders []*entity.Order
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&entity.Order{}).
		Where("status = ?", status).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error

	return orders, total, err
}

func (r *orderRepositoryImpl) GetOrdersByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entity.Order, int64, error) {
	var orders []*entity.Order
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&entity.Order{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error

	return orders, total, err
}

func (r *orderRepositoryImpl) GetOrdersByPaymentStatus(ctx context.Context, paymentStatus entity.PaymentStatus, limit, offset int) ([]*entity.Order, int64, error) {
	var orders []*entity.Order
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&entity.Order{}).
		Where("payment_status = ?", paymentStatus).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("payment_status = ?", paymentStatus).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error

	return orders, total, err
}

func (r *orderRepositoryImpl) GetOrderSummariesByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.OrderSummary, int64, error) {
	var summaries []*entity.OrderSummary
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&entity.Order{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated summaries
	err := r.db.WithContext(ctx).
		Select("id, order_number, status, payment_status, item_count, total_price, created_at, estimated_delivery").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&summaries).Error

	return summaries, total, err
}

func (r *orderRepositoryImpl) GetRecentOrders(ctx context.Context, limit int) ([]*entity.OrderSummary, error) {
	var summaries []*entity.OrderSummary

	err := r.db.WithContext(ctx).
		Select("id, order_number, status, payment_status, item_count, total_price, created_at, estimated_delivery").
		Order("created_at DESC").
		Limit(limit).
		Find(&summaries).Error

	return summaries, err
}

func (r *orderRepositoryImpl) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status entity.OrderStatus, notes string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if notes != "" {
		updates["notes"] = notes
	}

	// Set status-specific timestamps
	now := time.Now()
	switch status {
	case entity.OrderStatusConfirmed:
		updates["confirmed_at"] = now
	case entity.OrderStatusShipped:
		updates["shipped_at"] = now
	case entity.OrderStatusDelivered:
		updates["delivered_at"] = now
	case entity.OrderStatusCompleted:
		updates["completed_at"] = now
	case entity.OrderStatusCancelled:
		updates["cancelled_at"] = now
		if notes != "" {
			updates["cancel_reason"] = notes
		}
	}

	return r.db.WithContext(ctx).Model(&entity.Order{}).
		Where("id = ?", orderID).
		Updates(updates).Error
}

func (r *orderRepositoryImpl) UpdatePaymentStatus(ctx context.Context, orderID uuid.UUID, paymentStatus entity.PaymentStatus, transactionID string) error {
	updates := map[string]interface{}{
		"payment_status": paymentStatus,
		"updated_at":     time.Now(),
	}

	if transactionID != "" {
		updates["transaction_id"] = transactionID
	}

	return r.db.WithContext(ctx).Model(&entity.Order{}).
		Where("id = ?", orderID).
		Updates(updates).Error
}

func (r *orderRepositoryImpl) UpdateTrackingInfo(ctx context.Context, orderID uuid.UUID, trackingNumber string, estimatedDelivery *time.Time) error {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if trackingNumber != "" {
		updates["tracking_number"] = trackingNumber
	}

	if estimatedDelivery != nil {
		updates["estimated_delivery"] = estimatedDelivery
	}

	return r.db.WithContext(ctx).Model(&entity.Order{}).
		Where("id = ?", orderID).
		Updates(updates).Error
}

func (r *orderRepositoryImpl) GetOrderStatistics(ctx context.Context, startDate, endDate time.Time) (*repository.OrderStatistics, error) {
	var stats repository.OrderStatistics

	// Get basic statistics
	var totalOrders int64
	var totalRevenue float64

	if err := r.db.WithContext(ctx).Model(&entity.Order{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&totalOrders).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Model(&entity.Order{}).
		Where("created_at BETWEEN ? AND ? AND status != ?", startDate, endDate, entity.OrderStatusCancelled).
		Select("COALESCE(SUM(total_price), 0)").
		Scan(&totalRevenue).Error; err != nil {
		return nil, err
	}

	stats.TotalOrders = totalOrders
	stats.TotalRevenue = totalRevenue
	if totalOrders > 0 {
		stats.AverageOrderValue = totalRevenue / float64(totalOrders)
	}

	// Get status breakdown
	var statusResults []struct {
		Status entity.OrderStatus
		Count  int64
	}

	if err := r.db.WithContext(ctx).Model(&entity.Order{}).
		Select("status, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("status").
		Scan(&statusResults).Error; err != nil {
		return nil, err
	}

	stats.StatusBreakdown = make(map[entity.OrderStatus]int64)
	for _, result := range statusResults {
		stats.StatusBreakdown[result.Status] = result.Count
	}

	// Get payment status breakdown
	var paymentResults []struct {
		PaymentStatus entity.PaymentStatus
		Count         int64
	}

	if err := r.db.WithContext(ctx).Model(&entity.Order{}).
		Select("payment_status, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("payment_status").
		Scan(&paymentResults).Error; err != nil {
		return nil, err
	}

	stats.PaymentBreakdown = make(map[entity.PaymentStatus]int64)
	for _, result := range paymentResults {
		stats.PaymentBreakdown[result.PaymentStatus] = result.Count
	}

	// Get daily order statistics
	var dailyResults []struct {
		Date    time.Time
		Orders  int64
		Revenue float64
	}

	if err := r.db.WithContext(ctx).Model(&entity.Order{}).
		Select("DATE(created_at) as date, COUNT(*) as orders, COALESCE(SUM(total_price), 0) as revenue").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&dailyResults).Error; err != nil {
		return nil, err
	}

	stats.OrdersByDay = make([]repository.DailyOrderStats, len(dailyResults))
	for i, result := range dailyResults {
		stats.OrdersByDay[i] = repository.DailyOrderStats{
			Date:    result.Date,
			Orders:  result.Orders,
			Revenue: result.Revenue,
		}
	}

	return &stats, nil
}

func (r *orderRepositoryImpl) GetTopProducts(ctx context.Context, startDate, endDate time.Time, limit int) ([]*repository.ProductSalesInfo, error) {
	var products []*repository.ProductSalesInfo

	err := r.db.WithContext(ctx).
		Table("order_items oi").
		Select(`
			oi.product_id,
			oi.product_name,
			oi.product_sku,
			SUM(oi.quantity) as total_quantity,
			SUM(oi.total_price) as total_revenue,
			COUNT(DISTINCT oi.order_id) as order_count
		`).
		Joins("JOIN orders o ON oi.order_id = o.id").
		Where("o.created_at BETWEEN ? AND ? AND o.status != ?", startDate, endDate, entity.OrderStatusCancelled).
		Group("oi.product_id, oi.product_name, oi.product_sku").
		Order("total_quantity DESC").
		Limit(limit).
		Scan(&products).Error

	return products, err
}

func (r *orderRepositoryImpl) GetSalesMetrics(ctx context.Context, startDate, endDate time.Time) (*repository.SalesMetrics, error) {
	var metrics repository.SalesMetrics

	// Get basic metrics
	err := r.db.WithContext(ctx).
		Table("orders").
		Select(`
			COALESCE(SUM(total_price), 0) as total_revenue,
			COUNT(*) as total_orders,
			COALESCE(AVG(total_price), 0) as average_order_value,
			COALESCE(SUM(item_count), 0) as total_items
		`).
		Where("created_at BETWEEN ? AND ? AND status != ?", startDate, endDate, entity.OrderStatusCancelled).
		Scan(&metrics).Error

	if err != nil {
		return nil, err
	}

	if metrics.TotalOrders > 0 {
		metrics.AverageItemsPerOrder = float64(metrics.TotalItems) / float64(metrics.TotalOrders)
	}

	// Calculate conversion rate (assuming this would come from analytics data)
	// This would typically involve session/visitor data which we don't have in orders
	metrics.ConversionRate = 0.0

	// Calculate return rate
	var returnedOrders int64
	if err := r.db.WithContext(ctx).Model(&entity.Order{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entity.OrderStatusRefunded).
		Count(&returnedOrders).Error; err != nil {
		return nil, err
	}

	if metrics.TotalOrders > 0 {
		metrics.ReturnRate = float64(returnedOrders) / float64(metrics.TotalOrders) * 100
	}

	return &metrics, nil
}

func (r *orderRepositoryImpl) SearchOrders(ctx context.Context, filters *repository.OrderFilters) ([]*entity.Order, int64, error) {
	var orders []*entity.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Order{})

	// Apply filters
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.PaymentStatus != nil {
		query = query.Where("payment_status = ?", *filters.PaymentStatus)
	}

	if filters.DateFrom != nil {
		query = query.Where("created_at >= ?", *filters.DateFrom)
	}

	if filters.DateTo != nil {
		query = query.Where("created_at <= ?", *filters.DateTo)
	}

	if filters.MinAmount != nil {
		query = query.Where("total_price >= ?", *filters.MinAmount)
	}

	if filters.MaxAmount != nil {
		query = query.Where("total_price <= ?", *filters.MaxAmount)
	}

	if filters.SearchTerm != "" {
		searchPattern := "%" + filters.SearchTerm + "%"
		query = query.Where(`
			order_number ILIKE ? OR
			shipping_first_name ILIKE ? OR
			shipping_last_name ILIKE ? OR
			billing_first_name ILIKE ? OR
			billing_last_name ILIKE ?
		`, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderBy := "created_at DESC"
	if filters.SortBy != "" {
		direction := "DESC"
		if filters.SortOrder == "asc" {
			direction = "ASC"
		}
		orderBy = fmt.Sprintf("%s %s", filters.SortBy, direction)
	}

	// Get results with pagination
	err := query.
		Preload("Items").
		Order(orderBy).
		Limit(filters.Limit).
		Offset(filters.Offset).
		Find(&orders).Error

	return orders, total, err
}

// Helper function to generate order numbers
func (r *orderRepositoryImpl) generateOrderNumber() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("ORD-%d", timestamp)
}