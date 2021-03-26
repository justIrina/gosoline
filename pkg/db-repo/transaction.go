package db_repo

import (
	"context"
	"fmt"
	"github.com/applike/gosoline/pkg/cfg"
	"github.com/applike/gosoline/pkg/mon"
	"github.com/applike/gosoline/pkg/tracing"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const TransactionStateIdle = "idle"
const TransactionStateBegin = "begin"
const TransactionStateCommit = "commit"
const TransactionStateFinished = "done"

//go:generate mockery -name Transaction
type Transaction interface {
	AddAction(function func(ctx context.Context, value ModelBased) error, ctx context.Context, value ModelBased)
	Begin() error
	Commit() error
	IsInBeginState() bool
	GetDBConnection() *gorm.DB
}

type transaction struct {
	logger  mon.Logger
	tracer  tracing.Tracer
	orm     *gorm.DB
	actions []*transactionAction
	state   string
}

type transactionAction struct {
	function func(ctx context.Context, value ModelBased) error
	context  context.Context
	value    ModelBased
}

func NewTransaction(config cfg.Config, logger mon.Logger) (*transaction, error) {
	tracer, err := tracing.ProvideTracer(config, logger)
	if err != nil {
		return nil, fmt.Errorf("can not create tracer: %w", err)
	}

	orm, err := NewOrm(config, logger)
	if err != nil {
		return nil, fmt.Errorf("can not create orm: %w", err)
	}

	orm.Callback().
		Update().
		After("gorm:update_time_stamp").
		Register("gosoline:ignore_created_at_if_needed", ignoreCreatedAtIfNeeded)

	return NewTransactionWithInterfaces(logger, tracer, orm), nil
}

func NewTransactionWithInterfaces(logger mon.Logger, tracer tracing.Tracer, orm *gorm.DB) *transaction {
	return &transaction{
		logger: logger,
		tracer: tracer,
		orm:    orm,
		state:  TransactionStateIdle,
	}
}

func (t *transaction) AddAction(function func(ctx context.Context, value ModelBased) error, ctx context.Context, value ModelBased) {
	action := &transactionAction{
		function: function,
		context:  ctx,
		value:    value,
	}
	t.actions = append(t.actions, action)
}

func (t *transaction) Begin() error {
	if t.state != TransactionStateIdle {
		return errors.New("transaction start allowed only from idle state. Current state: " + t.state)
	}
	t.orm = t.orm.Begin()
	t.state = TransactionStateBegin

	return nil
}

func (t *transaction) Commit() error {
	if t.state != TransactionStateBegin {
		return errors.New("transaction commit allowed only from begin state. Current state: " + t.state)
	}

	t.state = TransactionStateCommit

	for _, action := range t.actions {
		err := action.function(action.context, action.value)
		if err != nil {
			_ = t.rollback()
			return err
		}
	}

	t.orm.Commit()
	t.state = TransactionStateFinished

	return nil
}

func (t *transaction) GetDBConnection() *gorm.DB {
	return t.orm
}

func (t *transaction) IsInBeginState() bool {
	return t.state == TransactionStateBegin
}

func (t *transaction) rollback() error {
	// todo not sure if this check is needed. Currently - it isn't, but for possible future changes it can stay as a hint
	if t.state != TransactionStateCommit {
		return errors.New("transaction rollback allowed only from commit state. Current state: " + t.state)
	}

	t.orm.Rollback()
	t.state = TransactionStateFinished

	return nil
}
