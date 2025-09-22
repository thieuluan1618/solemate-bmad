package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"solemate/services/payment-service/internal/domain/entity"
	"solemate/services/payment-service/internal/domain/repository"
)

type paymentRepositoryImpl struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) repository.PaymentRepository {
	return &paymentRepositoryImpl{db: db}
}

// Payment CRUD operations
func (r *paymentRepositoryImpl) CreatePayment(ctx context.Context, payment *entity.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *paymentRepositoryImpl) GetPaymentByID(ctx context.Context, paymentID uuid.UUID) (*entity.Payment, error) {
	var payment entity.Payment
	err := r.db.WithContext(ctx).
		Preload("PaymentMethod").
		Preload("Refunds").
		First(&payment, "id = ?", paymentID).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepositoryImpl) GetPaymentByStripePaymentIntentID(ctx context.Context, stripePaymentIntentID string) (*entity.Payment, error) {
	var payment entity.Payment
	err := r.db.WithContext(ctx).
		Preload("PaymentMethod").
		Preload("Refunds").
		First(&payment, "stripe_payment_intent_id = ?", stripePaymentIntentID).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepositoryImpl) GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*entity.Payment, error) {
	var payment entity.Payment
	err := r.db.WithContext(ctx).
		Preload("PaymentMethod").
		Preload("Refunds").
		First(&payment, "order_id = ?", orderID).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepositoryImpl) UpdatePayment(ctx context.Context, payment *entity.Payment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

func (r *paymentRepositoryImpl) DeletePayment(ctx context.Context, paymentID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Payment{}, "id = ?", paymentID).Error
}

// Payment querying
func (r *paymentRepositoryImpl) GetPaymentsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Payment, int64, error) {
	var payments []*entity.Payment
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.Payment{}).Where("user_id = ?", userID)

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("PaymentMethod").
		Preload("Refunds").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error

	return payments, count, err
}

func (r *paymentRepositoryImpl) GetPaymentsByStatus(ctx context.Context, status entity.PaymentStatus, limit, offset int) ([]*entity.Payment, int64, error) {
	var payments []*entity.Payment
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.Payment{}).Where("status = ?", status)

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("PaymentMethod").
		Preload("Refunds").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error

	return payments, count, err
}

func (r *paymentRepositoryImpl) GetPaymentsByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entity.Payment, int64, error) {
	var payments []*entity.Payment
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.Payment{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("PaymentMethod").
		Preload("Refunds").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error

	return payments, count, err
}

func (r *paymentRepositoryImpl) GetPaymentSummariesByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.PaymentSummary, int64, error) {
	var summaries []*entity.PaymentSummary
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.Payment{}).Where("user_id = ?", userID)

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	rows, err := query.
		Select("id, order_id, amount, currency, status, created_at, processed_at").
		Joins("LEFT JOIN payment_methods ON payments.payment_method_id = payment_methods.id").
		Select("payments.id, payments.order_id, payments.amount, payments.currency, payments.status, payments.created_at, payments.processed_at, payment_methods.card_last4, payment_methods.card_brand").
		Order("payments.created_at DESC").
		Limit(limit).
		Offset(offset).
		Rows()

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var summary entity.PaymentSummary
		var cardLast4, cardBrand *string

		err := rows.Scan(
			&summary.ID,
			&summary.OrderID,
			&summary.Amount,
			&summary.Currency,
			&summary.Status,
			&summary.CreatedAt,
			&summary.ProcessedAt,
			&cardLast4,
			&cardBrand,
		)
		if err != nil {
			return nil, 0, err
		}

		if cardLast4 != nil {
			summary.CardLast4 = *cardLast4
		}
		if cardBrand != nil {
			summary.CardBrand = *cardBrand
		}

		summaries = append(summaries, &summary)
	}

	return summaries, count, nil
}

// Payment analytics
func (r *paymentRepositoryImpl) GetPaymentStatistics(ctx context.Context, startDate, endDate time.Time) (*repository.PaymentStatistics, error) {
	var stats repository.PaymentStatistics

	// Get basic counts and totals
	err := r.db.WithContext(ctx).Model(&entity.Payment{}).
		Select("COUNT(*) as total_payments, SUM(amount) as total_revenue").
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	// Get successful payments count
	err = r.db.WithContext(ctx).Model(&entity.Payment{}).
		Where("status = ? AND created_at >= ? AND created_at <= ?", entity.PaymentStatusSucceeded, startDate, endDate).
		Count(&stats.SuccessfulPayments).Error
	if err != nil {
		return nil, err
	}

	// Get failed payments count
	err = r.db.WithContext(ctx).Model(&entity.Payment{}).
		Where("status = ? AND created_at >= ? AND created_at <= ?", entity.PaymentStatusFailed, startDate, endDate).
		Count(&stats.FailedPayments).Error
	if err != nil {
		return nil, err
	}

	// Get refunded amount
	err = r.db.WithContext(ctx).Model(&entity.Refund{}).
		Where("status = ? AND created_at >= ? AND created_at <= ?", entity.RefundStatusSucceeded, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.RefundedAmount).Error
	if err != nil {
		return nil, err
	}

	// Get status breakdown
	rows, err := r.db.WithContext(ctx).Model(&entity.Payment{}).
		Select("status, COUNT(*) as count").
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Group("status").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats.StatusBreakdown = make(map[entity.PaymentStatus]int64)
	for rows.Next() {
		var status entity.PaymentStatus
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats.StatusBreakdown[status] = count
	}

	// Get daily payment stats
	dailyRows, err := r.db.WithContext(ctx).Model(&entity.Payment{}).
		Select("DATE(created_at) as date, COUNT(*) as payments, SUM(amount) as revenue").
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("date").
		Rows()
	if err != nil {
		return nil, err
	}
	defer dailyRows.Close()

	for dailyRows.Next() {
		var daily repository.DailyPaymentStats
		if err := dailyRows.Scan(&daily.Date, &daily.Payments, &daily.Revenue); err != nil {
			return nil, err
		}
		stats.PaymentsByDay = append(stats.PaymentsByDay, daily)
	}

	return &stats, nil
}

func (r *paymentRepositoryImpl) GetPaymentMethodStats(ctx context.Context, startDate, endDate time.Time) ([]*repository.PaymentMethodStats, error) {
	var stats []*repository.PaymentMethodStats

	rows, err := r.db.WithContext(ctx).Table("payments").
		Select("payment_methods.type, payment_methods.card_brand as brand, COUNT(*) as count, SUM(payments.amount) as total_amount").
		Joins("LEFT JOIN payment_methods ON payments.payment_method_id = payment_methods.id").
		Where("payments.created_at >= ? AND payments.created_at <= ? AND payments.status = ?", startDate, endDate, entity.PaymentStatusSucceeded).
		Group("payment_methods.type, payment_methods.card_brand").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var stat repository.PaymentMethodStats
		var pmType *string
		var brand *string

		err := rows.Scan(&pmType, &brand, &stat.Count, &stat.TotalAmount)
		if err != nil {
			return nil, err
		}

		if pmType != nil {
			stat.Type = entity.PaymentMethodType(*pmType)
		}
		if brand != nil {
			stat.Brand = *brand
		}

		stats = append(stats, &stat)
	}

	return stats, nil
}

func (r *paymentRepositoryImpl) GetRevenueMetrics(ctx context.Context, startDate, endDate time.Time) (*repository.RevenueMetrics, error) {
	var metrics repository.RevenueMetrics

	// Get total revenue and transaction count
	err := r.db.WithContext(ctx).Model(&entity.Payment{}).
		Select("COALESCE(SUM(amount), 0) as total_revenue, COUNT(*) as transaction_count").
		Where("status = ? AND created_at >= ? AND created_at <= ?", entity.PaymentStatusSucceeded, startDate, endDate).
		Scan(&metrics).Error
	if err != nil {
		return nil, err
	}

	// Get refunded revenue
	err = r.db.WithContext(ctx).Model(&entity.Refund{}).
		Where("status = ? AND created_at >= ? AND created_at <= ?", entity.RefundStatusSucceeded, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&metrics.RefundedRevenue).Error
	if err != nil {
		return nil, err
	}

	// Calculate derived metrics
	metrics.NetRevenue = metrics.TotalRevenue - metrics.RefundedRevenue
	if metrics.TransactionCount > 0 {
		metrics.AverageTransaction = metrics.TotalRevenue / float64(metrics.TransactionCount)
	}
	if metrics.TotalRevenue > 0 {
		metrics.RefundRate = (metrics.RefundedRevenue / metrics.TotalRevenue) * 100
	}

	return &metrics, nil
}