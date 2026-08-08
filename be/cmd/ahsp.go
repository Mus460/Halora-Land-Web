package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/xuri/excelize/v2"

	"github.com/halora-land/halora-be/internal/ahsp"
	"github.com/halora-land/halora-be/internal/database"
	"github.com/halora-land/halora-be/internal/env"
)

// openFile opens an xlsx file for the import-ahsp CLI.
func openFile(path string) (*excelize.File, error) {
	return excelize.OpenFile(path)
}

func importAHSP(cfg *env.Config) error {
	fs := flag.NewFlagSet("import-ahsp", flag.ExitOnError)
	file := fs.String("file", DEFAULT_AHSP_PATH, "path to ahsp xlsx")
	sheet := fs.String("sheet", "", "sheet name (empty = all importable sheets)")
	force := fs.Bool("force", false, "delete existing sheet rows before reimport")
	_ = fs.Parse(os.Args[2:])

	ctx := context.Background()
	pool, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	im := ahsp.NewImporter(pool)
	f, err := openFile(*file)
	if err != nil {
		return fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()

	sheets := ahsp.ListSheets(f)
	if *sheet != "" {
		sheets = []string{*sheet}
	}

	prices, err := ahsp.ParsePriceList(f)
	if err != nil {
		return fmt.Errorf("parse price list: %w", err)
	}
	n, err := im.ImportPriceList(ctx, prices, *force)
	if err != nil {
		return fmt.Errorf("import price list: %w", err)
	}
	log.Printf("price-list: items=%d inserted=%d", len(prices), n)

	for _, sh := range sheets {
		items, err := ahsp.ParseSheet(f, sh)
		if err != nil {
			log.Printf("parse %s: %v", sh, err)
			continue
		}
		res, err := im.ImportSheet(ctx, items, prices, *force)
		if err != nil {
			log.Printf("import %s: %v", sh, err)
			continue
		}
		log.Printf("sheet=%s items=%d components=%d skipped=%d", res.Sheet, res.Items, res.Components, res.Skipped)
	}
	return nil
}
