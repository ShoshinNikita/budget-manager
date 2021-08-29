package api

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/reqid"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

type BackupHandlers struct {
	db  BackupDB
	log logger.Logger
}

type BackupDB interface {
	GetAllIncomes(context.Context) ([]db.Income, error)
	GetAllMonthlyPayments(context.Context) ([]db.MonthlyPayment, error)
	GetAllSpends(context.Context) ([]db.Spend, error)
	GetAllSpendTypes(context.Context) ([]db.SpendType, error)
}

// @Summary Download all data as a zip archive
// @Tags Backup
// @Router /api/backup/zip [get]
// @Produce application/zip
// @Success 200 {body} zip
// @Failure 500 {object} models.Response "Internal error"
//
func (h BackupHandlers) BackupZipArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	now := time.Now()

	backup, err := h.backupData(ctx)
	if err != nil {
		utils.EncodeInternalError(ctx, w, log, "couldn't backup data", err)
		return
	}

	zipArchive, err := prepareZipArchive(backup, now)
	if err != nil {
		utils.EncodeInternalError(ctx, w, log, "couldn't prepare zip archive", err)
		return
	}

	filename := "budget_manager_backup_" + now.Format("2006-01-02") + ".zip"

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	if _, err := io.Copy(w, zipArchive); err != nil {
		utils.LogInternalError(log, "couldn't write zip archive", err)
	}
}

func (h BackupHandlers) backupData(ctx context.Context) (backup models.Backup, err error) {
	wrap := func(err error, model string) error {
		return errors.Wrapf(err, "couldn't get all %s", model)
	}

	if backup.Incomes, err = h.db.GetAllIncomes(ctx); err != nil {
		return models.Backup{}, wrap(err, "incomes")
	}
	if backup.MonthlyPayments, err = h.db.GetAllMonthlyPayments(ctx); err != nil {
		return models.Backup{}, wrap(err, "monthly payments")
	}
	if backup.Spends, err = h.db.GetAllSpends(ctx); err != nil {
		return models.Backup{}, wrap(err, "spends")
	}
	if backup.SpendTypes, err = h.db.GetAllSpendTypes(ctx); err != nil {
		return models.Backup{}, wrap(err, "spend types")
	}
	return backup, nil
}

func prepareZipArchive(backup models.Backup, now time.Time) (io.Reader, error) {
	encodeJSON := func(w io.Writer, v interface{}) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}

	buf := &bytes.Buffer{}

	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	for _, b := range []struct {
		name string
		data interface{}
	}{
		{"incomes", backup.Incomes},
		{"monthly payments", backup.MonthlyPayments},
		{"spends", backup.Spends},
		{"spend types", backup.SpendTypes},
	} {
		header, err := zipWriter.CreateHeader(&zip.FileHeader{
			Name:     strings.ReplaceAll(b.name, " ", "_") + ".json",
			Method:   zip.Deflate,
			Modified: now,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't prepare zip header for %s", b.name)
		}

		if err := encodeJSON(header, b.data); err != nil {
			return nil, errors.Wrapf(err, "couldn't write %s to archive", b.name)
		}
	}

	return buf, nil
}

// @Summary Download all data as a JSON
// @Tags Backup
// @Router /api/backup/json [get]
// @Produce application/json
// @Success 200 {object} models.BackupResp
// @Failure 500 {object} models.Response "Internal error"
//
func (h BackupHandlers) BackupJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	backup, err := h.backupData(ctx)
	if err != nil {
		utils.EncodeInternalError(ctx, w, log, "couldn't backup data", err)
		return
	}

	resp := &models.BackupResp{
		Backup: backup,
	}
	utils.Encode(ctx, w, log, utils.EncodeResponse(resp))
}
