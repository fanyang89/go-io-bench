package main

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/gobwas/glob"
	"github.com/goccy/go-json"
	"github.com/iancoleman/strcase"
	"github.com/karlseguin/jsonwriter"
	"github.com/rodaine/table"
	"github.com/schwarmco/go-cartesian-product"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var cmd = &cli.Command{
	Name: "go-io-bench",
	Commands: []*cli.Command{
		cmdGenerate,
		cmdReadCsv,
	},
}

type scriptGenerator struct {
	taskName string
	fioPath  string
	fileName string
	size     string
	runtime  string
	w        io.Writer
}

func (g *scriptGenerator) Write() error {
	_, err := fmt.Fprintln(g.w, "#!/bin/bash")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(g.w, "set -x")
	if err != nil {
		return err
	}

	bsList := []interface{}{"256k", "128k", "16k", "8k", "4k"}
	rwList := []interface{}{"write", "read", "randwrite", "randread"}
	ioDepthList := []interface{}{128, 32, 1}

	for product := range cartesian.Iter(bsList, rwList, ioDepthList) {
		bs := product[0].(string)
		rw := product[1].(string)
		ioDepth := product[2].(int)
		name := fmt.Sprintf("%s_%s_%s_%d", g.taskName, bs, rw, ioDepth)

		params := []string{
			g.fioPath,
			"--name", name,
			"--filename", g.getFileName(),
			"--bs", bs,
			"--rw", rw,
			"--ioengine", "libaio",
			"--direct", "1",
			"--iodepth", fmt.Sprintf("%d", ioDepth),
			"--numjobs", "1",
			"--size", g.size,
			"--output-format", "csv",
			"--output", name + ".csv",
			"--time_based",
			"--runtime", g.runtime,
		}

		_, err = fmt.Fprintln(g.w, strings.Join(params, " "))
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *scriptGenerator) getFileName() string {
	if g.fileName != "" {
		return g.fileName
	}
	return fmt.Sprintf("%s_test-file", g.taskName)
}

//go:embed fio_minimal_header.csv
var fioCsvColumns string

var fioFieldNames = strings.FieldsFunc(strings.TrimSpace(fioCsvColumns), func(r rune) bool { return r == ',' })

func tryParse(s string) (interface{}, error) {
	u, err := strconv.ParseUint(s, 10, 32)
	if err == nil {
		return uint32(u), nil
	}

	u, err = strconv.ParseUint(s, 10, 64)
	if err == nil {
		return u, nil
	}

	i, err := strconv.ParseInt(s, 10, 32)
	if err == nil {
		return int32(i), nil
	}

	i, err = strconv.ParseInt(s, 10, 64)
	if err == nil {
		return i, nil
	}

	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return f, nil
	}

	return nil, errors.New("not a number")
}

func parseFioResults(path string, fileGlob glob.Glob) ([]FioResult, error) {
	results := make([]FioResult, 0)
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !fileGlob.Match(path) {
			return nil
		}

		buf, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		fields := strings.FieldsFunc(strings.TrimSpace(string(buf)), func(r rune) bool { return r == ';' })

		if len(fields) != len(fioFieldNames) {
			panic(errors.Newf("wrong number of fields, path: %s", path))
		}

		buffer := new(bytes.Buffer)
		w := jsonwriter.New(buffer)

		w.RootObject(func() {
			for i, fieldName := range fioFieldNames {
				fieldName = strcase.ToSnake(fieldName)
				field := fields[i]
				var fieldValue interface{}
				fieldValue, err = tryParse(field)
				if err != nil {
					w.KeyValue(fieldName, field)
				} else {
					w.KeyValue(fieldName, fieldValue)
				}
			}
		})

		decoder := json.NewDecoder(bufio.NewReader(buffer))
		decoder.DisallowUnknownFields()

		var result FioResult
		err = decoder.Decode(&result)
		if err != nil {
			return err
		}

		results = append(results, result)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func zeroToSlash(s string) string {
	if s == "0" || s == "0 B" || s == "0.000 us" {
		return "-"
	}
	return s
}

var cmdReadCsv = &cli.Command{
	Name: "read-csv",
	Arguments: []cli.Argument{
		&cli.StringArg{Name: "path", Config: cli.StringConfig{TrimSpace: true}},
		&cli.StringArg{Name: "glob", Config: cli.StringConfig{TrimSpace: true}},
	},
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "show-table", Aliases: []string{"t"}},
	},
	Action: func(ctx context.Context, command *cli.Command) error {
		showTable := command.Bool("show-table")
		path := command.StringArg("path")
		pathGlob := command.StringArg("glob")
		if pathGlob == "" {
			pathGlob = "*.csv"
		}

		g, err := glob.Compile(pathGlob)
		if err != nil {
			return err
		}

		if err != nil {
			return err
		}

		results, err := parseFioResults(path, g)
		if err != nil {
			return err
		}

		if showTable {
			slices.SortFunc(results, func(a, b FioResult) int {
				return strings.Compare(a.JobName, b.JobName)
			})

			tbl := table.New("Name",
				"Read IOPS", "Read bw", "Read lat",
				"Write IOPS", "Write bw ", "Write Lat")
			tbl.WithHeaderFormatter(color.New(color.FgGreen, color.Underline).SprintfFunc()).
				WithFirstColumnFormatter(color.New(color.FgYellow).SprintfFunc())

			for _, result := range results {
				tbl.AddRow(result.JobName,
					zeroToSlash(humanize.SI(humanize.ComputeSI(float64(result.ReadIOPS)))),
					zeroToSlash(humanize.Bytes(result.ReadBandwidth*1000)),
					zeroToSlash(fmt.Sprintf("%.3f us", result.ReadLatMean)),
					zeroToSlash(humanize.SI(humanize.ComputeSI(float64(result.WriteIOPS)))),
					zeroToSlash(humanize.Bytes(result.WriteBandwidth*1000)),
					zeroToSlash(fmt.Sprintf("%.3f us", result.WriteLatMean)),
				)
			}

			tbl.Print()
			return nil
		}

		// print JSONs
		for _, result := range results {
			buf, err := json.Marshal(result)
			if err != nil {
				return err
			}
			fmt.Println(string(buf))
		}
		return nil
	},
}

type FioResult struct {
	TerseVersion uint32 `json:"terse_version"`
	FioVersion   string `json:"fio_version"`
	JobName      string `json:"jobname"`
	GroupID      uint32 `json:"groupid"`
	Error        uint32 `json:"error"`

	ReadKB        uint64  `json:"read_kb"`
	ReadBandwidth uint64  `json:"read_bandwidth"`
	ReadIOPS      uint64  `json:"read_iops"`
	ReadRuntime   uint64  `json:"read_runtime"`
	ReadSlatMin   uint64  `json:"read_slat_min"`
	ReadSlatMax   uint64  `json:"read_slat_max"`
	ReadSlatMean  float64 `json:"read_slat_mean"`
	ReadSlatDev   float64 `json:"read_slat_dev"`
	ReadClatMax   uint64  `json:"read_clat_max"`
	ReadClatMin   uint64  `json:"read_clat_min"`
	ReadClatMean  float64 `json:"read_clat_mean"`
	ReadClatDev   float64 `json:"read_clat_dev"`
	ReadClatPct01 string  `json:"read_clat_pct_01"`
	ReadClatPct02 string  `json:"read_clat_pct_02"`
	ReadClatPct03 string  `json:"read_clat_pct_03"`
	ReadClatPct04 string  `json:"read_clat_pct_04"`
	ReadClatPct05 string  `json:"read_clat_pct_05"`
	ReadClatPct06 string  `json:"read_clat_pct_06"`
	ReadClatPct07 string  `json:"read_clat_pct_07"`
	ReadClatPct08 string  `json:"read_clat_pct_08"`
	ReadClatPct09 string  `json:"read_clat_pct_09"`
	ReadClatPct10 string  `json:"read_clat_pct_10"`
	ReadClatPct11 string  `json:"read_clat_pct_11"`
	ReadClatPct12 string  `json:"read_clat_pct_12"`
	ReadClatPct13 string  `json:"read_clat_pct_13"`
	ReadClatPct14 string  `json:"read_clat_pct_14"`
	ReadClatPct15 string  `json:"read_clat_pct_15"`
	ReadClatPct16 string  `json:"read_clat_pct_16"`
	ReadClatPct17 string  `json:"read_clat_pct_17"`
	ReadClatPct18 string  `json:"read_clat_pct_18"`
	ReadClatPct19 string  `json:"read_clat_pct_19"`
	ReadClatPct20 string  `json:"read_clat_pct_20"`
	ReadTlatMin   uint64  `json:"read_tlat_min"`
	ReadLatMax    uint64  `json:"read_lat_max"`
	ReadLatMean   float64 `json:"read_lat_mean"`
	ReadLatDev    float64 `json:"read_lat_dev"`
	ReadBWMin     uint64  `json:"read_bw_min"`
	ReadBWMax     uint64  `json:"read_bw_max"`
	ReadBWAggPct  string  `json:"read_bw_agg_pct"`
	ReadBWMean    float64 `json:"read_bw_mean"`
	ReadBWDev     float64 `json:"read_bw_dev"`

	WriteKB        uint64  `json:"write_kb"`
	WriteBandwidth uint64  `json:"write_bandwidth"`
	WriteIOPS      uint64  `json:"write_iops"`
	WriteRuntime   uint64  `json:"write_runtime"`
	WriteSlatMin   uint64  `json:"write_slat_min"`
	WriteSlatMax   uint64  `json:"write_slat_max"`
	WriteSlatMean  float64 `json:"write_slat_mean"`
	WriteSlatDev   float64 `json:"write_slat_dev"`
	WriteClatMax   uint64  `json:"write_clat_max"`
	WriteClatMin   uint64  `json:"write_clat_min"`
	WriteClatMean  float64 `json:"write_clat_mean"`
	WriteClatDev   float64 `json:"write_clat_dev"`
	WriteClatPct01 string  `json:"write_clat_pct_01"`
	WriteClatPct02 string  `json:"write_clat_pct_02"`
	WriteClatPct03 string  `json:"write_clat_pct_03"`
	WriteClatPct04 string  `json:"write_clat_pct_04"`
	WriteClatPct05 string  `json:"write_clat_pct_05"`
	WriteClatPct06 string  `json:"write_clat_pct_06"`
	WriteClatPct07 string  `json:"write_clat_pct_07"`
	WriteClatPct08 string  `json:"write_clat_pct_08"`
	WriteClatPct09 string  `json:"write_clat_pct_09"`
	WriteClatPct10 string  `json:"write_clat_pct_10"`
	WriteClatPct11 string  `json:"write_clat_pct_11"`
	WriteClatPct12 string  `json:"write_clat_pct_12"`
	WriteClatPct13 string  `json:"write_clat_pct_13"`
	WriteClatPct14 string  `json:"write_clat_pct_14"`
	WriteClatPct15 string  `json:"write_clat_pct_15"`
	WriteClatPct16 string  `json:"write_clat_pct_16"`
	WriteClatPct17 string  `json:"write_clat_pct_17"`
	WriteClatPct18 string  `json:"write_clat_pct_18"`
	WriteClatPct19 string  `json:"write_clat_pct_19"`
	WriteClatPct20 string  `json:"write_clat_pct_20"`
	WriteTlatMin   uint64  `json:"write_tlat_min"`
	WriteLatMax    uint64  `json:"write_lat_max"`
	WriteLatMean   float64 `json:"write_lat_mean"`
	WriteLatDev    float64 `json:"write_lat_dev"`
	WriteBWMin     uint64  `json:"write_bw_min"`
	WriteBWMax     uint64  `json:"write_bw_max"`
	WriteBWAggPct  string  `json:"write_bw_agg_pct"`
	WriteBWMean    float64 `json:"write_bw_mean"`
	WriteBWDev     float64 `json:"write_bw_dev"`

	CpuUser string `json:"cpu_user"`
	CpuSys  string `json:"cpu_sys"`
	CpuCsw  uint64 `json:"cpu_csw"`
	CpuMjf  uint64 `json:"cpu_mjf"`
	PuMinf  uint64 `json:"pu_minf"`

	IODepth1  string `json:"iodepth_1"`
	IODepth2  string `json:"iodepth_2"`
	IODepth4  string `json:"iodepth_4"`
	IODepth8  string `json:"iodepth_8"`
	IODepth16 string `json:"iodepth_16"`
	IODepth32 string `json:"iodepth_32"`
	IODepth64 string `json:"iodepth_64"`

	Lat2us        string `json:"lat_2_us"`
	Lat4us        string `json:"lat_4_us"`
	Lat10us       string `json:"lat_10_us"`
	Lat20us       string `json:"lat_20_us"`
	Lat50us       string `json:"lat_50_us"`
	Lat100us      string `json:"lat_100_us"`
	Lat250us      string `json:"lat_250_us"`
	Lat500us      string `json:"lat_500_us"`
	Lat750us      string `json:"lat_750_us"`
	Lat1000us     string `json:"lat_1000_us"`
	Lat2ms        string `json:"lat_2_ms"`
	Lat4ms        string `json:"lat_4_ms"`
	Lat10ms       string `json:"lat_10_ms"`
	Lat20ms       string `json:"lat_20_ms"`
	Lat50ms       string `json:"lat_50_ms"`
	Lat100ms      string `json:"lat_100_ms"`
	Lat250ms      string `json:"lat_250_ms"`
	Lat500ms      string `json:"lat_500_ms"`
	Lat750ms      string `json:"lat_750_ms"`
	Lat1000ms     string `json:"lat_1000_ms"`
	Lat2000ms     string `json:"lat_2000_ms"`
	LatOver2000ms string `json:"lat_over_2000_ms"`

	DiskName        string `json:"disk_name"`
	DiskReadIOPS    uint64 `json:"disk_read_iops"`
	DiskWriteIOPS   uint64 `json:"disk_write_iops"`
	DiskReadMerges  uint64 `json:"disk_read_merges"`
	DiskWriteMerges uint64 `json:"disk_write_merges"`
	DiskReadTicks   uint64 `json:"disk_read_ticks"`
	WriteTicks      uint64 `json:"write_ticks"`
	DiskQueueTime   uint64 `json:"disk_queue_time"`
	DiskUtilization string `json:"disk_utilization"`
}

var cmdGenerate = &cli.Command{
	Name:    "generate",
	Aliases: []string{"gen", "g"},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "fio-path",
			Value: "fio",
		},
		&cli.StringFlag{
			Name:     "task-name",
			Aliases:  []string{"n", "task", "name"},
			Required: true,
		},
		&cli.StringFlag{
			Name: "file-name",
		},
		&cli.StringFlag{
			Name:  "size",
			Value: "10G",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Value:   "-",
		},
		&cli.StringFlag{
			Name:    "runtime",
			Aliases: []string{"r"},
			Value:   "60s",
		},
	},
	Action: func(ctx context.Context, command *cli.Command) error {
		outputPath := command.String("output")
		var w io.Writer
		if outputPath == "-" {
			w = os.Stdout
		} else {
			f, err := os.Create(outputPath)
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()
			w = f
		}

		g := &scriptGenerator{
			taskName: command.String("task-name"),
			fioPath:  command.String("fio-path"),
			fileName: command.String("file-name"),
			size:     command.String("size"),
			runtime:  command.String("runtime"),
			w:        w,
		}
		return g.Write()
	},
}

func main() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := config.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	defer func() { _ = logger.Sync() }()

	err = cmd.Run(context.Background(), os.Args)
	if err != nil {
		zap.L().Error("Unexpected error", zap.Error(err))
		os.Exit(1)
	}
}
